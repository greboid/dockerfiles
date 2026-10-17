#define _POSIX_C_SOURCE 200809L
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/resource.h>
#include <unistd.h>

static void hex(const char *value) {
    for (const unsigned char *p = (const unsigned char *)value; *p; p++) printf("%02x", *p);
    putchar('\n');
}

int main(int argc, char **argv) {
    struct rlimit limit;
    if (getrlimit(RLIMIT_CORE, &limit)) return 2;
    printf("PID %ld\nUID %ld\nCORE %llu %llu\n", (long)getpid(), (long)getuid(),
           (unsigned long long)limit.rlim_cur, (unsigned long long)limit.rlim_max);
    for (int i = 0; i < argc; i++) { fputs("ARG ", stdout); hex(argv[i]); }
    const char *names[] = {"SE_ONE", "SE_TWO", "SE_KEEP", NULL};
    for (int i = 0; names[i]; i++) {
        const char *value = getenv(names[i]);
        if (value) { printf("ENV %s ", names[i]); hex(value); }
    }
    puts("END");
    fflush(stdout);
    for (int i = 1; i < argc; i++) {
        if (!strcmp(argv[i], "--hold")) for (;;) pause();
        if (!strcmp(argv[i], "--exit=23")) return 23;
    }
    return 0;
}
