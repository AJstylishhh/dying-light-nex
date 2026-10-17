#include <switch.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <sys/stat.h>

#define DL_HOST "g27332e00-lp1.s.n.srv.nintendo.net"
#define HOSTS_SYSMMC "sdmc:/atmosphere/hosts/sysmmc.txt"
#define HOSTS_EMUMMC "sdmc:/atmosphere/hosts/emummc.txt"
#define MAX_FILE_SIZE (128 * 1024)

static void pause_screen(void) {
    printf("\nPress A to continue...\n");
    while (appletMainLoop()) {
        hidScanInput();
        if (hidKeysDown(CONTROLLER_P1_AUTO) & KEY_A) break;
        consoleUpdate(NULL);
    }
}

static int valid_ipv4(const char *s) {
    int a, b, c, d;
    char tail;
    if (sscanf(s, "%d.%d.%d.%d%c", &a, &b, &c, &d, &tail) != 4) return 0;
    return a >= 0 && a <= 255 && b >= 0 && b <= 255 &&
           c >= 0 && c <= 255 && d >= 0 && d <= 255;
}

static char *read_file_or_empty(const char *path) {
    FILE *f = fopen(path, "rb");
    if (!f) return calloc(1, 1);
    if (fseek(f, 0, SEEK_END) != 0) { fclose(f); return NULL; }
    long n = ftell(f);
    if (n < 0 || n > MAX_FILE_SIZE || fseek(f, 0, SEEK_SET) != 0) {
        fclose(f); return NULL;
    }
    char *buf = calloc((size_t)n + 1, 1);
    if (!buf) { fclose(f); return NULL; }
    size_t got = fread(buf, 1, (size_t)n, f);
    fclose(f);
    if (got != (size_t)n) { free(buf); return NULL; }
    return buf;
}

static int line_is_dl_host(const char *line) {
    char copy[512];
    size_t n = strcspn(line, "\r\n");
    if (n >= sizeof(copy)) return 0;
    memcpy(copy, line, n);
    copy[n] = '\0';

    char *p = copy;
    while (isspace((unsigned char)*p)) p++;
    if (*p == '#' || *p == '\0') return 0;

    char ip[64] = {0}, host[256] = {0};
    if (sscanf(p, "%63s %255s", ip, host) != 2) return 0;
    return strcmp(host, DL_HOST) == 0;
}

static int write_hosts(const char *path, const char *ip, int apply) {
    char *old = read_file_or_empty(path);
    if (!old) return 0;

    size_t cap = strlen(old) + 512;
    char *out = calloc(cap, 1);
    if (!out) { free(old); return 0; }
    size_t used = 0;
    int found = 0;

    char *cursor = old;
    while (*cursor) {
        char *end = strchr(cursor, '\n');
        size_t len = end ? (size_t)(end - cursor + 1) : strlen(cursor);
        char line[512];
        size_t copylen = len < sizeof(line) - 1 ? len : sizeof(line) - 1;
        memcpy(line, cursor, copylen);
        line[copylen] = '\0';

        if (line_is_dl_host(line)) {
            found = 1;
            /* Replace existing Dying Light entries with exactly one new rule. */
            if (apply && !strstr(out, DL_HOST)) {
                int n = snprintf(out + used, cap - used, "%s %s\n", ip, DL_HOST);
                if (n > 0 && (size_t)n < cap - used) used += (size_t)n;
            }
        } else {
            if (used + len + 1 >= cap) {
                size_t newcap = cap + len + 512;
                char *grown = realloc(out, newcap);
                if (!grown) { free(old); free(out); return 0; }
                out = grown; cap = newcap;
            }
            memcpy(out + used, cursor, len);
            used += len;
            out[used] = '\0';
        }
        cursor += len;
    }

    if (apply && !found) {
        int n = snprintf(out + used, cap - used, "%s %s\n", ip, DL_HOST);
        if (n < 0 || (size_t)n >= cap - used) {
            free(old); free(out); return 0;
        }
        used += (size_t)n;
    }

    FILE *f = fopen(path, "wb");
    if (!f) { free(old); free(out); return 0; }
    int ok = fwrite(out, 1, used, f) == used && fflush(f) == 0;
    fclose(f);
    free(old);
    free(out);
    return ok;
}

static int ensure_hosts_dir(void) {
    mkdir("sdmc:/atmosphere", 0777);
    mkdir("sdmc:/atmosphere/hosts", 0777);
    return 1;
}

static void show_keyboard_ip(char *out, size_t out_size) {
    SwkbdConfig kbd;
    swkbdCreate(&kbd, 0);
    swkbdConfigMakePresetDefault(&kbd);
    swkbdConfigSetHeaderText(&kbd, "Dying Light server IPv4");
    swkbdConfigSetSubText(&kbd, "Enter the reachable server IP");
    swkbdConfigSetType(&kbd, SwkbdType_Normal);
    swkbdConfigSetStringLenMax(&kbd, 15);
    swkbdConfigSetInitialText(&kbd, "192.168.137.1");
    swkbdShow(&kbd, out, out_size);
    swkbdClose(&kbd);
}

int main(int argc, char **argv) {
    (void)argc; (void)argv;
    consoleInit(NULL);
    printf("Dying Light Prelude - DNS redirector\n");
    printf("------------------------------------\n");
    printf("Host: %s\n\n", DL_HOST);
    printf("This utility only edits Atmosphere host files.\n");
    printf("It does not run the NEX server.\n\n");
    printf("X = Apply redirect\n");
    printf("Y = Remove redirect\n");
    printf("B = Exit\n");

    while (appletMainLoop()) {
        hidScanInput();
        u64 k = hidKeysDown(CONTROLLER_P1_AUTO);
        if (k & KEY_B) break;

        if (k & KEY_X) {
            char ip[64] = {0};
            show_keyboard_ip(ip, sizeof(ip));
            if (!valid_ipv4(ip)) {
                printf("\nInvalid IPv4 address: %s\n", ip);
                pause_screen();
            } else {
                ensure_hosts_dir();
                int a = write_hosts(HOSTS_SYSMMC, ip, 1);
                int b = write_hosts(HOSTS_EMUMMC, ip, 1);
                if (a && b) {
                    fsdevCommitDevice("sdmc");
                    printf("\nRedirect applied to sysMMC + emuMMC.\n");
                    printf("IP: %s\n", ip);
                    printf("Reboot the Switch to test.\n");
                } else {
                    printf("\nCould not write one or both host files.\n");
                    printf("Check SD card and Atmosphere paths.\n");
                }
                pause_screen();
                printf("\nDying Light Prelude - DNS redirector\n");
                printf("X = Apply redirect | Y = Remove | B = Exit\n");
            }
        }

        if (k & KEY_Y) {
            int a = write_hosts(HOSTS_SYSMMC, "", 0);
            int b = write_hosts(HOSTS_EMUMMC, "", 0);
            if (a && b) {
                fsdevCommitDevice("sdmc");
                printf("\nDying Light hostname removed from both host files.\n");
            } else {
                printf("\nCould not update one or both host files.\n");
            }
            pause_screen();
            printf("\nDying Light Prelude - DNS redirector\n");
            printf("X = Apply redirect | Y = Remove | B = Exit\n");
        }

        consoleUpdate(NULL);
    }

    consoleExit(NULL);
    return 0;
}
