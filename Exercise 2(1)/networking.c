#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>

// #define PORT 30000
// #define BUFFER_SIZE 1024

// int main(void) {
//     int sockfd;
//     struct sockaddr_in server_addr, client_addr;
//     char buffer[BUFFER_SIZE];
//     socklen_t addr_len = sizeof(client_addr);

//     // Create socket
//     if ((sockfd = socket(AF_INET, SOCK_DGRAM, 0)) < 0) {
//         perror("socket creation failed");
//         exit(EXIT_FAILURE);
//     }

//     // Set socket options to allow broadcast
//     int broadcast = 1;
//     if (setsockopt(sockfd, SOL_SOCKET, SO_BROADCAST, &broadcast, sizeof(broadcast)) < 0) {
//         perror("setsockopt failed");
//         close(sockfd);
//         exit(EXIT_FAILURE);
//     }

//     // Bind the socket to the port
//     memset(&server_addr, 0, sizeof(server_addr));
//     server_addr.sin_family = AF_INET;
//     server_addr.sin_addr.s_addr = INADDR_ANY;
//     server_addr.sin_port = htons(PORT);

//     if (bind(sockfd, (const struct sockaddr *)&server_addr, sizeof(server_addr)) < 0) {
//         perror("bind failed");
//         close(sockfd);
//         exit(EXIT_FAILURE);
//     }

//     printf("Listening for broadcast on port %d\n", PORT);

//     // Listen for broadcast messages
//     while (1) {
//         int n = recvfrom(sockfd, buffer, BUFFER_SIZE, 0, (struct sockaddr *)&client_addr, &addr_len);
//         if (n < 0) {
//             perror("r30000ecvfrom failed");
//             close(sockfd);
//             exit(EXIT_FAILURE);
//         }
//         buffer[n] = '\0';
//         printf("Received broadcast message: %s\n", buffer);
//     }

//     close(sockfd);
//     return 0;
// }

#define PORT 30000
#define BUFFER_SIZE 1024


void reciever() {
    int sockfd;

    struct sockaddr_in sock_adr, client_adr, server_adr;
    char buffer[BUFFER_SIZE];
    socklen_t addr_len = sizeof(client_adr);



      sockfd = socket(AF_INET, SOCK_DGRAM, 0); 

       int broadcast = 1;   

    if(setsockopt(sockfd, SOL_SOCKET, SO_BROADCAST, &broadcast, sizeof(broadcast))<0)
    {
        perror("setsockopt failed");
        close(sockfd);
        exit(EXIT_FAILURE);
    }

    memset(&server_adr, 0, sizeof(server_adr));
    server_adr.sin_port = htons(PORT);
    server_adr.sin_family = AF_INET;
    server_adr.sin_addr.s_addr = INADDR_ANY;

    if (bind(sockfd, (struct sockaddr *)&server_adr, sizeof(server_adr)) <0 )
    {
        perror("bind failed");
        close(sockfd);
        exit(EXIT_FAILURE);
    }


    while(1)
    {
        memset(buffer, 0, BUFFER_SIZE);
        int n = recvfrom(sockfd, buffer, BUFFER_SIZE, 0, (struct sockaddr*)&client_adr, &addr_len);
        buffer[n] = '\0';
        if(client_adr.sin_addr.s_addr != server_adr.sin_addr.s_addr){
        printf("Received broadcast message: %s\n", buffer);
        sleep(0.5);
    }
    }

close(sockfd);
}
#define SEND_PORT 20007

void sender(){
    int sockfd;
    struct sockaddr_in sock_adr;
    char buffer[BUFFER_SIZE];
    socklen_t addr_len = sizeof(sock_adr);

    sockfd = socket(AF_INET, SOCK_DGRAM, 0);

    memset(&sock_adr, 0, sizeof(sock_adr));
    sock_adr.sin_family = AF_INET;
    sock_adr.sin_port = htons(SEND_PORT);
    sock_adr.sin_addr.s_addr = INADDR_BROADCAST;

    while(1){
        printf("Sending broadcast message\n");
        char *message = "Hello Server";
        sendto(sockfd, message, strlen(message), 0, (struct sockaddr*)&sock_adr, sizeof(sock_adr));
        sleep(1);
    }

    close(sockfd);
  
}


int main(void)
{
    
    sender();

    return 0;
}




