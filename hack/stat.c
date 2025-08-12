#include <stdio.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <unistd.h>

int main(int argc, char *argv[]) {
    if (argc != 2) {
        fprintf(stderr, "Usage: %s <path>\n", argv[0]);
        return 1;
    }

    struct stat st;
    if (stat(argv[1], &st) != 0) {
        perror("stat");
        return 1;
    }

    printf("File: %s\n", argv[1]);
    printf("Device ID (st_dev): %lu (0x%lx)\n", (unsigned long) st.st_dev, (unsigned long) st.st_dev);
    printf("Inode: %lu\n", (unsigned long) st.st_ino);

    return 0;
}
