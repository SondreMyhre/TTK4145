// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>
#include <semaphore.h>

int i = 0;
sem_t sem;

// Note the return type: void*
void* incrementingThreadFunction(){
    // TODO: increment i 1_000_000 times
    for(int j=0; j<1000000; j++){
        sem_wait(&sem);
        i++;
        sem_post(&sem);
    }
    return NULL;
}

void* decrementingThreadFunction(){
    // TODO: decrement i 1_000_000 times
    // sem_post(&sem);      Problem med semaphore er at hvilkensomhelst tråd kan poste, mens med mutex har vi eierskap
    for(int j=0; j<1000000; j++){
        sem_wait(&sem);
        i--;
        sem_post(&sem);
    }
    return NULL;
}


int main(){
    // TODO: 
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?
    pthread_t t1, t2;
    sem_init(&sem, 0, 1);
    pthread_create(&t1, NULL, incrementingThreadFunction, NULL);
    pthread_create(&t2, NULL, decrementingThreadFunction, NULL);

    
    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`    

    pthread_join(t1, NULL);
    pthread_join(t2, NULL);
    sem_destroy(&sem);
    
    printf("The magic number is: %d\n", i);
    return 0;
}
