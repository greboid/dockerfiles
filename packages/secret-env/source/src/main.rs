#[cfg(not(target_os = "linux"))]
compile_error!("secret-env requires Linux");

use std::env;
use std::ffi::{CString, OsStr, OsString};
use std::fs::File;
use std::io::{self, Read};
use std::os::fd::{AsRawFd, FromRawFd, OwnedFd};
use std::os::unix::ffi::{OsStrExt, OsStringExt};
use std::os::unix::fs::MetadataExt;

const LIMIT: usize = 64 * 1024;
const MAX_SECRETS: usize = 128;
const CONFIG_DIR: &str = "/etc/secret-env";
type Result<T> = std::result::Result<T, &'static str>;

#[derive(Debug)]
struct Config {
    path: String,
    secrets: Vec<(String, String)>,
}

fn valid_path(path: &[u8]) -> bool {
    path.len() > 1
        && path.len() < 4096
        && path[0] == b'/'
        && !path.iter().any(|b| *b < 32 || *b == 127)
        && path[1..]
            .split(|b| *b == b'/')
            .all(|part| !part.is_empty() && part.len() <= 255 && part != b"." && part != b"..")
}

fn valid_name(name: &str) -> bool {
    let b = name.as_bytes();
    !b.is_empty()
        && b.len() <= 255
        && (b[0].is_ascii_alphabetic() || b[0] == b'_')
        && b.iter().all(|c| c.is_ascii_alphanumeric() || *c == b'_')
}

fn valid_app(name: &[u8]) -> bool {
    !name.is_empty()
        && name.len() <= 128
        && (name[0].is_ascii_alphanumeric() || name[0] == b'_')
        && name
            .iter()
            .all(|c| c.is_ascii_alphanumeric() || b"_.-".contains(c))
}

fn parse_config(bytes: &[u8]) -> Result<Config> {
    let text = std::str::from_utf8(bytes).map_err(|_| "invalid configuration encoding")?;
    if bytes
        .iter()
        .any(|b| (*b < 32 && !b"\n\r\t".contains(b)) || *b == 127)
    {
        return Err("invalid configuration character");
    }
    let mut section = "";
    let mut seen_exec = false;
    let mut seen_secrets = false;
    let mut path = None;
    let mut secrets = Vec::new();
    for line in text.split('\n') {
        let line = line.trim_matches([' ', '\t', '\r']);
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        match line {
            "[exec]" if !seen_exec => {
                section = "exec";
                seen_exec = true;
                continue;
            }
            "[secrets]" if !seen_secrets => {
                section = "secrets";
                seen_secrets = true;
                continue;
            }
            _ if line.starts_with('[') => return Err("invalid configuration section"),
            _ => {}
        }
        let (key, value) = line.split_once('=').ok_or("invalid configuration entry")?;
        let key = key.trim_matches([' ', '\t']);
        let value = value.trim_matches([' ', '\t']);
        if !valid_path(value.as_bytes()) {
            return Err("invalid configuration path");
        }
        match section {
            "exec" if key == "path" && path.is_none() => path = Some(value.to_owned()),
            "secrets" if valid_name(key) => {
                if secrets.len() == MAX_SECRETS || secrets.iter().any(|(name, _)| name == key) {
                    return Err("duplicate or excessive secret mapping");
                }
                secrets.push((key.to_owned(), value.to_owned()));
            }
            _ => return Err("invalid configuration key"),
        }
    }
    Ok(Config {
        path: path.ok_or("missing executable path")?,
        secrets,
    })
}

fn openat(dir: i32, name: &std::ffi::CStr, flags: i32) -> Result<OwnedFd> {
    loop {
        // name is NUL-terminated; no creation flags, so no mode argument is required.
        let fd = unsafe { libc::openat(dir, name.as_ptr(), flags) };
        if fd >= 0 {
            // A successful openat transfers this new descriptor to OwnedFd exactly once.
            return Ok(unsafe { OwnedFd::from_raw_fd(fd) });
        }
        if io::Error::last_os_error().kind() != io::ErrorKind::Interrupted {
            return Err("cannot open input file");
        }
    }
}

fn open_file(path: &OsStr, config: bool) -> Result<File> {
    let bytes = path.as_bytes();
    if !valid_path(bytes) {
        return Err("invalid input path");
    }
    let directory = libc::O_PATH | libc::O_DIRECTORY | libc::O_NOFOLLOW | libc::O_CLOEXEC;
    let mut fd = openat(libc::AT_FDCWD, c"/", directory)?;
    let mut parts = bytes[1..].split(|b| *b == b'/').peekable();
    while let Some(part) = parts.next() {
        let part = CString::new(part).map_err(|_| "invalid input path")?;
        let flags = if parts.peek().is_some() {
            directory
        } else {
            libc::O_RDONLY | libc::O_NOFOLLOW | libc::O_CLOEXEC | libc::O_NONBLOCK
        };
        fd = openat(fd.as_raw_fd(), &part, flags)?;
    }
    let file = File::from(fd);
    let metadata = file.metadata().map_err(|_| "cannot inspect input file")?;
    if !metadata.is_file() {
        return Err("input must be a regular file");
    }
    // geteuid has no preconditions and cannot fail.
    let uid = unsafe { libc::geteuid() };
    if config && ((metadata.uid() != 0 && metadata.uid() != uid) || metadata.mode() & 0o022 != 0) {
        return Err("unsafe configuration ownership or permissions");
    }
    Ok(file)
}

fn read_bounded(mut file: impl Read, prefix: &[u8]) -> Result<Vec<u8>> {
    let mut buf = Vec::new();
    buf.try_reserve_exact(prefix.len() + LIMIT + 1)
        .map_err(|_| "allocation failed")?;
    buf.extend_from_slice(prefix);
    buf.resize(prefix.len() + LIMIT + 1, 0);
    let mut end = prefix.len();
    loop {
        match file.read(&mut buf[end..]) {
            Ok(0) => break,
            Ok(n) => {
                end += n;
                if end - prefix.len() > LIMIT {
                    return Err("input exceeds size limit");
                }
            }
            Err(e) if e.kind() == io::ErrorKind::Interrupted => continue,
            Err(_) => return Err("cannot read input file"),
        }
    }
    buf.truncate(end);
    Ok(buf)
}

fn harden() -> Result<()> {
    // These Linux APIs take no borrowed pointers except the initialized rlimit.
    unsafe {
        if libc::getuid() != libc::geteuid() || libc::getgid() != libc::getegid() {
            return Err("privileged invocation is unsupported");
        }
        let limit = libc::rlimit {
            rlim_cur: 0,
            rlim_max: 0,
        };
        if libc::setrlimit(libc::RLIMIT_CORE, &limit) != 0
            || libc::prctl(
                libc::PR_SET_DUMPABLE,
                0 as libc::c_ulong,
                0 as libc::c_ulong,
                0 as libc::c_ulong,
                0 as libc::c_ulong,
            ) != 0
        {
            return Err("cannot apply process hardening");
        }
    }
    Ok(())
}

fn invocation(args: &[OsString]) -> Result<(OsString, Vec<OsString>, bool)> {
    let original = args.first().ok_or("missing invocation name")?;
    let name = original
        .as_bytes()
        .rsplit(|b| *b == b'/')
        .next()
        .unwrap_or_default();
    if name == b"secret-env" {
        if args.len() < 4 || args[1] != "--config" || args[3] != "--" {
            return Err("usage: secret-env --config /absolute/app.conf -- [arguments]");
        }
        Ok((args[2].clone(), args[4..].to_vec(), false))
    } else {
        if !valid_app(name) {
            return Err("invalid invocation name");
        }
        let name = std::str::from_utf8(name).map_err(|_| "invalid invocation name")?;
        Ok((
            format!("{CONFIG_DIR}/{name}.conf").into(),
            args[1..].to_vec(),
            true,
        ))
    }
}

fn run() -> Result<()> {
    harden()?;
    let args: Vec<OsString> = env::args_os().collect();
    let (config_path, app_args, native) = invocation(&args)?;
    let config = parse_config(&read_bounded(open_file(&config_path, true)?, b"")?)?;
    let target = CString::new(config.path.as_bytes()).map_err(|_| "invalid executable path")?;
    let mut argv = Vec::new();
    let name = if native {
        args[0].clone()
    } else {
        OsString::from(
            config
                .path
                .rsplit('/')
                .next()
                .ok_or("invalid executable path")?,
        )
    };
    for arg in std::iter::once(name).chain(app_args) {
        argv.push(CString::new(arg.into_vec()).map_err(|_| "invalid argument")?);
    }
    let mut environment = Vec::new();
    for (key, value) in env::vars_os() {
        if config
            .secrets
            .iter()
            .any(|(name, _)| key.as_bytes() == name.as_bytes())
        {
            continue;
        }
        let mut entry = key.into_vec();
        entry.push(b'=');
        entry.extend(value.into_vec());
        environment.push(CString::new(entry).map_err(|_| "invalid inherited environment")?);
    }
    for (name, path) in &config.secrets {
        let prefix = format!("{name}=");
        let mut value = read_bounded(open_file(OsStr::new(path), false)?, prefix.as_bytes())?;
        if value.contains(&0) {
            return Err("secret contains a NUL byte");
        }
        value.push(0);
        environment.push(CString::from_vec_with_nul(value).map_err(|_| "invalid secret value")?);
    }
    let mut argv_ptrs: Vec<_> = argv.iter().map(|s| s.as_ptr()).collect();
    argv_ptrs.push(std::ptr::null());
    let mut env_ptrs: Vec<_> = environment.iter().map(|s| s.as_ptr()).collect();
    env_ptrs.push(std::ptr::null());
    let error = loop {
        // Both vectors are terminated; CString storage stays live through execve.
        unsafe { libc::execve(target.as_ptr(), argv_ptrs.as_ptr(), env_ptrs.as_ptr()) };
        let error = io::Error::last_os_error();
        if error.kind() != io::ErrorKind::Interrupted {
            break error;
        }
    };
    Err(match error.raw_os_error() {
        Some(libc::E2BIG) => "exec failed: argument/environment limit exceeded",
        Some(libc::ENOENT) => "exec failed: executable or interpreter missing",
        Some(libc::EACCES) => "exec failed: permission denied",
        Some(libc::ENOEXEC) => "exec failed: unsupported executable format",
        _ => "exec failed",
    })
}

fn main() {
    if let Err(message) = run() {
        eprintln!("secret-env: {message}");
        std::process::exit(1);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn grammar() {
        assert!(
            parse_config(b"# comment\n[exec]\npath = /bin/app\n[secrets]\nA = /run/a\n").is_ok()
        );
        for text in [
            "",
            "[exec]\npath=/a\npath=/b",
            "[exec]\npath=/a\n[exec]",
            "[exec]\nother=/a",
            "[exec]\npath=relative",
            "[exec]\npath=\"/a\"",
            "[exec]\npath=/a\n[secrets]\nA=/a\nA=/b",
            "[exec]\npath=/a\n[unknown]",
            "[exec]\npath=/a\n[secrets]\n1A=/a",
            "[exec]\npath=/a\0",
        ] {
            assert!(parse_config(text.as_bytes()).is_err());
        }
    }

    #[test]
    fn bounded_short_and_interrupted_reads() {
        struct ShortRead {
            interrupted: bool,
            remaining: usize,
        }
        impl Read for ShortRead {
            fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
                self.interrupted = !self.interrupted;
                if self.interrupted {
                    return Err(io::ErrorKind::Interrupted.into());
                }
                let n = self.remaining.min(buf.len()).min(7);
                buf[..n].fill(b'x');
                self.remaining -= n;
                Ok(n)
            }
        }
        for length in [0, 1, LIMIT, LIMIT + 1] {
            let result = read_bounded(
                ShortRead {
                    interrupted: false,
                    remaining: length,
                },
                b"KEY=",
            );
            if length > LIMIT {
                assert!(result.is_err());
            } else {
                let result = result.unwrap();
                assert_eq!(result.len(), length + 4);
                assert_eq!(&result[..4], b"KEY=");
                assert!(result[4..].iter().all(|b| *b == b'x'));
            }
        }
    }

    #[test]
    fn mapping_limit() {
        let mut text = String::from("[exec]\npath=/bin/app\n[secrets]\n");
        for n in 0..MAX_SECRETS {
            text.push_str(&format!("KEY{n}=/run/secret\n"));
        }
        assert!(parse_config(text.as_bytes()).is_ok());
        text.push_str("EXTRA=/run/secret\n");
        assert!(parse_config(text.as_bytes()).is_err());
    }

    #[test]
    fn arbitrary_parser_input_does_not_panic() {
        let mut state = 1_u64;
        for length in 0..2048 {
            let mut bytes = vec![0; length];
            for byte in &mut bytes {
                state = state.wrapping_mul(6364136223846793005).wrapping_add(1);
                *byte = (state >> 32) as u8;
            }
            let _ = parse_config(&bytes);
        }
    }

    #[test]
    fn names_and_paths() {
        for path in [
            b"/".as_slice(),
            b"relative",
            b"/a/..",
            b"/a/./b",
            b"/a//b",
            b"/a/",
        ] {
            assert!(!valid_path(path));
        }
        for name in [b"..".as_slice(), b"../app", b".app", b"a/b", b""] {
            assert!(!valid_app(name));
        }
        assert!(valid_app(b"app-name_1.0"));
        assert!(valid_name("_TOKEN1"));
        assert!(!valid_name("TOKEN-1"));
    }
}
