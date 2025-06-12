#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <windows.h>
#include <openssl/evp.h>
#include <openssl/err.h>

#define MAX_PATH_LEN 1024
#define BUFFER_SIZE 4096
#define MAX_THREADS 8
#define MAX_FILES   100000

typedef struct {
    char filepath[MAX_PATH_LEN];
} FileTask;

typedef struct {
    FileTask* tasks;
    int task_count;
    int current_index;
    CRITICAL_SECTION lock;
    FILE* out_file;
    int processed;
} ThreadPool;

// 计算单个文件的 MD5 值
int compute_md5(const char* path, char* md5_out) {
    FILE* file = fopen(path, "rb");
    if (!file) return -1;

    EVP_MD_CTX* ctx = EVP_MD_CTX_new();
    if (!ctx) {
        fclose(file);
        return -1;
    }

    if (!EVP_DigestInit_ex(ctx, EVP_md5(), NULL)) {
        EVP_MD_CTX_free(ctx);
        fclose(file);
        return -1;
    }

    unsigned char data[BUFFER_SIZE];
    unsigned char digest[EVP_MAX_MD_SIZE];
    unsigned int digest_len;
    size_t bytes;

    while ((bytes = fread(data, 1, BUFFER_SIZE, file)) > 0) {
        if (!EVP_DigestUpdate(ctx, data, bytes)) {
            EVP_MD_CTX_free(ctx);
            fclose(file);
            return -1;
        }
    }

    if (!EVP_DigestFinal_ex(ctx, digest, &digest_len)) {
        EVP_MD_CTX_free(ctx);
        fclose(file);
        return -1;
    }

    EVP_MD_CTX_free(ctx);
    fclose(file);

    for (unsigned int i = 0; i < digest_len; i++) {
        sprintf(md5_out + i * 2, "%02x", digest[i]);
    }
    md5_out[32] = '\0';
    return 0;
}

DWORD WINAPI worker_thread(LPVOID arg) {
    ThreadPool* pool = (ThreadPool*)arg;
    while (1) {
        int index;

        EnterCriticalSection(&pool->lock);
        if (pool->current_index >= pool->task_count) {
            LeaveCriticalSection(&pool->lock);
            break;
        }
        index = pool->current_index++;
        LeaveCriticalSection(&pool->lock);

        char md5[33];
        const char* path = pool->tasks[index].filepath;
        int ok = compute_md5(path, md5);

        EnterCriticalSection(&pool->lock);
        if (ok == 0) {
            fprintf(pool->out_file, "%s  %s\n", path, md5);
        } else {
            fprintf(pool->out_file, "%s  ERROR\n", path);
        }
        pool->processed++;
        if (pool->processed % 100 == 0 || pool->processed == pool->task_count) {
            printf("进度: %.1f%% (%d/%d)\n", (float)pool->processed / pool->task_count * 100, pool->processed, pool->task_count);
        }
        LeaveCriticalSection(&pool->lock);
    }
    return 0;
}

int main() {
    SetConsoleOutputCP(CP_UTF8);
    FILE* list = fopen("output.txt", "r");
    if (!list) {
        fprintf(stderr, "无法打开 output.txt\n");
        return 1;
    }

    FileTask* tasks = malloc(MAX_FILES * sizeof(FileTask));
    int count = 0;
    while (fgets(tasks[count].filepath, MAX_PATH_LEN, list)) {
        tasks[count].filepath[strcspn(tasks[count].filepath, "\r\n")] = '\0';
        count++;
        if (count >= MAX_FILES) break;
    }
    fclose(list);

    FILE* out = fopen("md5_output.txt", "w");
    if (!out) {
        fprintf(stderr, "无法创建输出文件\n");
        free(tasks);
        return 1;
    }

    ThreadPool pool = {
        .tasks = tasks,
        .task_count = count,
        .current_index = 0,
        .out_file = out,
        .processed = 0
    };
    InitializeCriticalSection(&pool.lock);

    HANDLE threads[MAX_THREADS];
    int threads_to_create = min(MAX_THREADS, count);
    for (int i = 0; i < threads_to_create; i++) {
        threads[i] = CreateThread(NULL, 0, worker_thread, &pool, 0, NULL);
    }

    WaitForMultipleObjects(threads_to_create, threads, TRUE, INFINITE);
    for (int i = 0; i < threads_to_create; i++) {
        CloseHandle(threads[i]);
    }

    DeleteCriticalSection(&pool.lock);
    fclose(out);
    free(tasks);
    printf("所有文件已处理，结果保存在 md5_output.txt\n");
    return 0;
} 