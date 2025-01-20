

#include <assert.h>
#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <termios.h>
#include <unistd.h>
#include <string.h>
 
#include "arduino_io_card.h"



SerialLine openSerial(char* port, int reentrant){
    int r = 0;
    
    int handle = open(port, (O_RDWR | O_NOCTTY));
    if(handle < 0){
        printf("open() error: %s\n", strerror(errno));
        assert(!"Unable to open serial line");
    }
    
    struct termios tty;
    r = tcgetattr(handle, &tty);
    assert(r >= 0);
    
    tty.c_cflag = (B9600 | CLOCAL | CREAD | CS8);
    tty.c_lflag = 0;
    tty.c_iflag = 0;
    tty.c_oflag = 0;
    tty.c_cc[VMIN] = 0;
    tty.c_cc[VTIME] = 10;   // 10 tenths of a second to respond after trying to read from the device
    
    tcflush(handle, TCIFLUSH);
    r = tcsetattr(handle, TCSANOW, &tty);
    assert(r >= 0);

    
    SerialLine s = {
        .handle     = handle,
        .reentrant  = reentrant,
        .mutex      = PTHREAD_MUTEX_INITIALIZER,
    };
    
    // For some unknown reason, the device does not respond immediately the first time it is connected.
    // This load-bearing sleep prevents the program from timing out during the device response check just below.
    usleep(100*1000);  
    
    uint8_t v = readPins(&s, 63);
    if(v != 0xa2){
        printf("Received 0x%02x instead of 0xa2", v);
        assert(!"Serial line is not responding as an NTNU Arduino IO Card");
    }
    
    return s;
}

char* autoDetectArduino(void){
    FILE* exec = popen("readlink -f $(find /dev/serial/by-id/ -name '*NTNU_Arduino_IO_Card*' | head -n 1)", "r");
    if(!exec){
        return NULL;
    }
    char* buf = calloc(144, sizeof(char));
    fgets(buf, 144, exec);
    size_t idx = strcspn(buf, "\r\n");
    if(!idx){
        return NULL;
    }
    buf[idx] = 0;
    return buf;
}

void writeByteImpl(SerialLine* const serial, uint8_t byte){
    int r = write(serial->handle, &byte, 1);    
    if(r < 0){
        printf("Error when writing to serial line (error: %s)\n", strerror(errno));
        assert(0);
    } else if(r == 0){
        assert(!"Unable to write to serial line.");
    }
}

uint8_t readByteImpl(SerialLine* const serial){
    uint8_t v;
    int r = read(serial->handle, &v, 1);
    if(r < 0){
        printf("Error when reading serial line (error: %s)\n", strerror(errno));
        assert(0);
    } else if(r == 0){
        assert(!"No data from serial line before timing out. Device did not respond to request-to-read");
    }
    return v;
}

void writePort(SerialLine* const serial, uint8_t port, uint8_t high){
    if(high){
        port |= 0b01000000;
    }
    
    if(serial->reentrant) pthread_mutex_lock(&serial->mutex);
    writeByteImpl(serial, port);
    if(serial->reentrant) pthread_mutex_unlock(&serial->mutex);
}

void writePWM(SerialLine* const serial, uint8_t port, uint8_t width){
    if(serial->reentrant) pthread_mutex_lock(&serial->mutex);
    writeByteImpl(serial, port);
    writeByteImpl(serial, width);
    if(serial->reentrant) pthread_mutex_unlock(&serial->mutex);
}

uint8_t readPin(SerialLine* const serial, uint8_t pin){
    pin |= 0b10000000;
    
    if(serial->reentrant) pthread_mutex_lock(&serial->mutex);
    writeByteImpl(serial, pin);
    uint8_t v = readByteImpl(serial);
    if(serial->reentrant) pthread_mutex_unlock(&serial->mutex);

    return !!v;
}

uint8_t readPins(SerialLine* const serial, uint8_t startingPin){
    startingPin |= 0b11000000;
    
    if(serial->reentrant) pthread_mutex_lock(&serial->mutex);
    writeByteImpl(serial, startingPin);
    uint8_t v = readByteImpl(serial);
    if(serial->reentrant) pthread_mutex_unlock(&serial->mutex);
    
    return v;
}
