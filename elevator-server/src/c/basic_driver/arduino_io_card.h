
#pragma once

#include <inttypes.h>
#include <pthread.h>

typedef struct SerialLine SerialLine;
struct SerialLine {
    int             handle;
    int             reentrant;
    pthread_mutex_t mutex;
};

SerialLine openSerial(char* port, int reentrant);
char* autoDetectArduino(void);
void writePort(SerialLine* const serial, uint8_t port, uint8_t high);
void writePWM(SerialLine* const serial, uint8_t port, uint8_t width);
uint8_t readPin(SerialLine* const serial, uint8_t pin);
uint8_t readPins(SerialLine* const serial, uint8_t startingPin);


//// Elevator ports and pins

#define PORT_LIGHT_UP0      30
#define PORT_LIGHT_UP1      31
#define PORT_LIGHT_UP2      29
#define PORT_LIGHT_UP3      0
#define PORT_LIGHT_DOWN0    0
#define PORT_LIGHT_DOWN1    28
#define PORT_LIGHT_DOWN2    26
#define PORT_LIGHT_DOWN3    27
#define PORT_LIGHT_CAB0     34
#define PORT_LIGHT_CAB1     35
#define PORT_LIGHT_CAB2     32
#define PORT_LIGHT_CAB3     33
#define PORT_LIGHT_STOP     37
#define PORT_FLOOR_HI       23
#define PORT_FLOOR_LO       22
#define PORT_DOOR           24
#define PORT_MOTORDIR       36
#define PWMPORT_MOTORSPEED  13

#define PIN_BTN_UP0         46
#define PIN_BTN_UP1         47
#define PIN_BTN_UP2         45
#define PIN_BTN_UP3         0
#define PIN_BTN_DOWN0       0
#define PIN_BTN_DOWN1       44
#define PIN_BTN_DOWN2       42
#define PIN_BTN_DOWN3       43
#define PIN_BTN_CAB0        50
#define PIN_BTN_CAB1        51
#define PIN_BTN_CAB2        48
#define PIN_BTN_CAB3        49
#define PIN_BTN_STOP        53
#define PIN_OBSTR           52
#define PIN_FLOOR0          40
#define PIN_FLOOR1          41
#define PIN_FLOOR2          38
#define PIN_FLOOR3          39
#define PIN_TACHO           62


//// Response Time ports and pins

#define PORT_RESPA          55
#define PORT_RESPB          57
#define PORT_RESPC          59

#define PIN_TESTA           54
#define PIN_TESTB           56      
#define PIN_TESTC           58


//// Wide pins

// Response time pin, returns pins [54 56 58]
#define PINS_RT     54
    #define PINS_RT_MASKA   0x01
    #define PINS_RT_MASKB   0x04
    #define PINS_RT_MASKC   0x10

// Wide pins connected to elevator
#define PINS_38     38
#define PINS_42     42
#define PINS_46     46

// Serial port ID response - reading this pin should always return 0xa2
#define PINS_ID     63