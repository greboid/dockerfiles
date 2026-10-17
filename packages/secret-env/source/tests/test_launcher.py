import os
from pathlib import Path
import shutil
import signal
import select
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
BINARY = Path(os.environ.get("SECRET_ENV_BIN", ROOT / "target/release/secret-env")).absolute()


def decode(output):
    result = {"args": [], "env": {}}
    for line in output.splitlines():
        key, _, data = line.partition(b" ")
        if key == b"ARG":
            result["args"].append(bytes.fromhex(data.decode()))
        elif key == b"ENV":
            name, _, value = data.partition(b" ")
            result["env"][name.decode()] = bytes.fromhex(value.decode())
        elif key in (b"PID", b"UID"):
            result[key.decode()] = int(data)
        elif key == b"CORE":
            result["CORE"] = data
    return result


class LauncherTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not BINARY.is_file():
            raise RuntimeError("build the launcher first: cargo build --release --locked")
        cls.build = tempfile.TemporaryDirectory(prefix="secret-env-probe-", dir=os.environ.get("TMPDIR"))
        cls.probe = Path(cls.build.name) / "probe.real"
        subprocess.run([os.environ.get("CC", "cc"), "-std=c11", "-Wall", "-Wextra", "-Werror",
                        str(ROOT / "tests/probe.c"), "-o", str(cls.probe)], check=True)

    @classmethod
    def tearDownClass(cls):
        cls.build.cleanup()

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="secret-env-test-", dir=os.environ.get("TMPDIR"))
        self.addCleanup(self.temporary.cleanup)
        self.directory = Path(self.temporary.name)
        self.config = self.directory / "app.conf"
        self.secret = self.directory / "secret"
        self.secret.write_bytes(b"synthetic-private-value-001")
        self.secret.chmod(0o600)
        self.configure()
        self.environment = os.environ.copy()
        for name in ("SE_ONE", "SE_TWO", "SE_KEEP"):
            self.environment.pop(name, None)

    def configure(self, mappings=None, target=None):
        if mappings is None:
            mappings = {"SE_ONE": self.secret}
        self.config.write_text("[exec]\npath = " + str(target or self.probe) + "\n[secrets]\n" +
                               "".join(f"{key} = {value}\n" for key, value in mappings.items()))
        self.config.chmod(0o600)

    def command(self, *args):
        return [str(BINARY), "--config", str(self.config), "--", *args]

    def launch(self, *args, command=None):
        return subprocess.run(command or self.command(*args), env=self.environment,
                              capture_output=True, timeout=5)

    def good(self, *args):
        result = self.launch(*args)
        self.assertEqual(result.returncode, 0, "launcher/target failed")
        self.assertEqual(result.stderr, b"")
        return decode(result.stdout)

    def bad(self, command=None):
        result = self.launch(command=command)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, b"")
        self.assertNotIn(b"synthetic-private-value-001", result.stderr)
        self.assertTrue(result.stderr.startswith(b"secret-env: "))
        return result

    def test_one_secret(self):
        self.assertEqual(self.good()["env"]["SE_ONE"], self.secret.read_bytes())

    def test_multiple_secrets(self):
        second = self.directory / "second"
        second.write_bytes(b"second-synthetic-value")
        self.configure({"SE_ONE": self.secret, "SE_TWO": second})
        self.assertEqual(self.good()["env"], {"SE_ONE": self.secret.read_bytes(), "SE_TWO": second.read_bytes()})

    def test_empty_secret(self):
        self.secret.write_bytes(b"")
        self.assertEqual(self.good()["env"]["SE_ONE"], b"")

    def test_trailing_newlines_preserved(self):
        for value in (b"value\n", b"value\r\n", b"value\n\n"):
            with self.subTest(length=len(value)):
                self.secret.write_bytes(value)
                self.assertEqual(self.good()["env"]["SE_ONE"], value)

    def test_non_utf8_secret(self):
        self.secret.write_bytes(b"\xff\x80\n")
        self.assertEqual(self.good()["env"]["SE_ONE"], b"\xff\x80\n")

    def test_nul_secret(self):
        self.secret.write_bytes(b"synthetic-private-value-001\0trailer")
        self.bad()

    def test_size_limit(self):
        self.secret.write_bytes(b"s" * 65536)
        self.assertEqual(len(self.good()["env"]["SE_ONE"]), 65536)
        self.secret.write_bytes(b"s" * 65537)
        self.bad()

    def test_kernel_environment_limit_fails_closed(self):
        self.secret.write_bytes(b"s" * 65536)
        self.configure({f"LIMIT{n}": self.secret for n in range(128)})
        self.assertIn(b"argument/environment limit exceeded", self.bad().stderr)

    def test_missing_config(self):
        self.config.unlink()
        self.bad()

    def test_invalid_configurations(self):
        cases = (
            (b"", b"missing executable path"),
            (b"[unknown]\n", b"invalid configuration section"),
            (b"[exec]\npath=relative\n", b"invalid configuration path"),
            (b"\xff", b"invalid configuration encoding"),
            (b"[exec]\npath=/a\0", b"invalid configuration character"),
            (b"[exec]\npath=/a\npath=/b", b"invalid configuration key"),
            (b"[exec]\npath=/a\n[exec]", b"invalid configuration section"),
            (b"[exec]\npath=/a\nextra=/b", b"invalid configuration key"),
            (b"[exec]\npath=\"/a\"", b"invalid configuration path"),
            (b"[exec]\npath=/a\n[secrets]\nA=/x\nA=/y", b"duplicate or excessive secret mapping"),
        )
        for value, diagnostic in cases:
            with self.subTest(config_length=len(value)):
                self.config.write_bytes(value)
                self.assertIn(diagnostic, self.bad().stderr)

    def test_config_size_limit(self):
        valid = self.config.read_bytes()
        self.config.write_bytes(valid + b"#" * (65537 - len(valid)))
        self.assertIn(b"input exceeds size limit", self.bad().stderr)

    def test_invalid_names(self):
        for name in ("", "1NAME", "A-B", "A.B", "NAME SPACE", "\u00e9", "X" * 256):
            with self.subTest(name_length=len(name)):
                self.configure({name: self.secret})
                self.bad()

    def test_invalid_paths(self):
        for path in ("relative", "/", "/run/../secret", "/run/./secret", "/run//secret", "/run/secret/"):
            self.configure({"SE_ONE": path})
            self.assertIn(b"invalid configuration path", self.bad().stderr)

    def test_missing_secret(self):
        self.secret.unlink()
        self.bad()

    def test_unreadable_secret(self):
        if os.geteuid() == 0:
            self.skipTest("permission-denied behavior requires non-root")
        self.secret.chmod(0)
        try:
            with self.secret.open("rb"):
                self.skipTest("test identity can bypass file permissions (e.g. build-guest capabilities)")
        except PermissionError:
            pass
        self.bad()

    def test_directory_and_fifo_rejected_without_blocking(self):
        self.secret.unlink()
        self.secret.mkdir()
        self.bad()
        self.secret.rmdir()
        os.mkfifo(self.secret)
        self.bad()

    def test_config_directory_and_fifo_rejected(self):
        self.config.unlink()
        self.config.mkdir()
        self.bad()
        self.config.rmdir()
        os.mkfifo(self.config)
        self.bad()

    def test_final_symlinks_rejected(self):
        real = self.directory / "real-secret"
        self.secret.rename(real)
        self.secret.symlink_to(real)
        self.bad()
        self.secret.unlink()
        real.rename(self.secret)
        real = self.directory / "real-config"
        self.config.rename(real)
        self.config.symlink_to(real)
        self.bad()

    def test_search_only_directory(self):
        directory = self.directory / "search-only"
        directory.mkdir()
        secret = directory / "secret"
        secret.write_bytes(b"search-only-directory-value")
        directory.chmod(0o100)
        try:
            self.configure({"SE_ONE": secret})
            self.assertEqual(self.good()["env"]["SE_ONE"], b"search-only-directory-value")
        finally:
            directory.chmod(0o700)

    def test_parent_symlinks_rejected(self):
        alias = self.directory / "alias"
        alias.symlink_to(self.directory, target_is_directory=True)
        self.configure({"SE_ONE": alias / "secret"})
        self.bad()
        self.configure()
        self.bad(command=[str(BINARY), "--config", str(alias / "app.conf"), "--"])

    def test_config_permissions(self):
        for mode in (0o620, 0o602, 0o666):
            self.config.chmod(mode)
            self.bad()

    def test_environment_override_and_preservation(self):
        self.environment.update(SE_ONE="old-value", SE_KEEP="preserved-value")
        values = self.good()["env"]
        self.assertEqual(values["SE_ONE"], self.secret.read_bytes())
        self.assertEqual(values["SE_KEEP"], b"preserved-value")

    def test_no_secrets_allowed(self):
        self.configure({})
        self.environment["SE_KEEP"] = "preserved-value"
        self.assertEqual(self.good()["env"], {"SE_KEEP": b"preserved-value"})

    def test_arguments_and_explicit_argv0(self):
        args = ["", "with spaces", "--config", "/not/a/config", "--", "$(literal)", "*.txt", "\u00e9"]
        self.assertEqual(self.good(*args)["args"], [b"probe.real", *[arg.encode() for arg in args]])

    def test_parser_requires_delimiter(self):
        self.bad(command=[str(BINARY), "--config", str(self.config)])
        self.bad(command=[str(BINARY), "--config", str(self.config), "--help"])
        self.bad(command=[str(BINARY), "--unknown", str(self.config), "--"])

    def test_failed_exec(self):
        self.configure(target=self.directory / "absent")
        self.bad()
        target = self.directory / "target"
        target.write_text("not a binary or a shebang script")
        self.configure(target=target)
        self.bad()
        target.chmod(0o700)
        result = self.bad()
        self.assertIn(b"unsupported executable format", result.stderr)

    def test_nonroot(self):
        if os.geteuid() == 0:
            self.skipTest("run suite as non-root to test non-root execution")
        self.assertEqual(self.good()["UID"], os.geteuid())

    def test_core_limits_survive_exec(self):
        self.assertEqual(self.good()["CORE"], b"0 0")

    def check_running_process(self, inspect_proc):
        process = subprocess.Popen(self.command("--hold"), env=self.environment,
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE, bufsize=0)
        try:
            output = b""
            while not output.endswith(b"END\n"):
                self.assertTrue(select.select([process.stdout], [], [], 5)[0], "target readiness timeout")
                line = process.stdout.readline()
                self.assertTrue(line, "target did not report readiness")
                output += line
            self.assertEqual(decode(output)["PID"], process.pid)
            if inspect_proc:
                command = Path(f"/proc/{process.pid}/cmdline").read_bytes()
                self.assertNotIn(self.secret.read_bytes(), command)
                self.assertTrue(os.path.samefile(f"/proc/{process.pid}/exe", self.probe))
            process.send_signal(signal.SIGTERM)
            _, errors = process.communicate(timeout=5)
            self.assertEqual(process.returncode, -signal.SIGTERM)
            self.assertEqual(errors, b"")
        finally:
            if process.poll() is None:
                process.kill()
                process.wait()
            process.stdout.close()
            process.stderr.close()

    def test_pid_replacement_and_signal(self):
        self.check_running_process(inspect_proc=False)

    def test_process_cmdline_and_executable(self):
        if not Path("/proc/self").exists() or os.readlink("/proc/self") != str(os.getpid()):
            self.skipTest("proc mount does not represent this PID namespace")
        self.check_running_process(inspect_proc=True)

    def test_target_exit_status(self):
        self.assertEqual(self.launch("--exit=23").returncode, 23)

    def test_failure_logs_after_first_secret_loaded(self):
        self.configure({"SE_ONE": self.secret, "SE_TWO": self.directory / "absent"})
        self.bad()

    def test_malicious_argv0(self):
        for name in ("..", ".hidden", "bad name", "x" * 129):
            result = subprocess.run([name], executable=BINARY, capture_output=True, timeout=5)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn(b"invalid invocation name", result.stderr)

    def test_symlink_invocation(self):
        if not shutil.which("bwrap"):
            self.skipTest("bwrap required for isolated /etc/secret-env")
        availability = subprocess.run(["bwrap", "--unshare-user", "--ro-bind", "/", "/", "--", "/bin/true"],
                                      capture_output=True, timeout=5)
        if availability.returncode:
            self.skipTest("user/mount namespaces unavailable")
        native = self.directory / "native"
        native.symlink_to(BINARY)
        command = ["bwrap", "--unshare-user", "--ro-bind", "/", "/", "--tmpfs", "/etc",
                   "--dir", "/etc/secret-env", "--ro-bind", str(self.config),
                   "/etc/secret-env/native.conf", "--", str(native), "--config", "--", "with spaces"]
        result = self.launch(command=command)
        self.assertEqual(result.returncode, 0, "native invocation failed")
        data = decode(result.stdout)
        self.assertEqual(data["args"], [os.fsencode(native), b"--config", b"--", b"with spaces"])
        self.assertEqual(data["env"]["SE_ONE"], self.secret.read_bytes())


if __name__ == "__main__":
    unittest.main()
