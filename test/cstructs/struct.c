#include <stdio.h>
#include <unistd.h>

typedef struct InnerStruct {
    int inner_int;
    double inner_double;
} InnerStruct_t;

struct OuterStruct {
    int outer_int;
    char outer_char;
    InnerStruct_t inner;
} OuterStruct_t;

struct OuterStruct global_struct = {
    .outer_int = 42,
    .outer_char = 'A',
    .inner = {
        .inner_int = 123,
        .inner_double = 3.14159
    }
};

int main() {
    printf("Global struct address: %p\n", &global_struct);
    printf("PID: %d\n", getpid());
    while (1) {
        sleep(1); // Keep the program running
    }
    return 0;
}