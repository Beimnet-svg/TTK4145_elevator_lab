#include <assert.h>
#include <stdlib.h>
#include <stdio.h>
#include <pthread.h>

#include "elevio.h"
#include "arduino_io_card.h"

static SerialLine serialLine;

void elevio_init(void){
    
    serialLine = openSerial("/dev/ttyACM0", 1);
    
}


static const uint8_t motorDrivePWM = 255;
static const uint8_t motorStopPWM  = 0;
void elevio_motorDirection(MotorDirection dirn){
    if(dirn == 0){
        writePWM(&serialLine, PWMPORT_MOTORSPEED, motorStopPWM);
    } else if(dirn > 0){
        writePort(&serialLine, PORT_MOTORDIR, 0);
        writePWM(&serialLine, PWMPORT_MOTORSPEED, motorDrivePWM);
    } else {
        writePort(&serialLine, PORT_MOTORDIR, 1);
        writePWM(&serialLine, PWMPORT_MOTORSPEED, motorDrivePWM);
    }
}

static const uint8_t btnLights[N_FLOORS][N_BUTTONS] = {
    {PORT_LIGHT_UP0, PORT_LIGHT_DOWN0, PORT_LIGHT_CAB0},
    {PORT_LIGHT_UP1, PORT_LIGHT_DOWN1, PORT_LIGHT_CAB1},
    {PORT_LIGHT_UP2, PORT_LIGHT_DOWN2, PORT_LIGHT_CAB2},
    {PORT_LIGHT_UP3, PORT_LIGHT_DOWN3, PORT_LIGHT_CAB3},
};
void elevio_buttonLamp(int floor, ButtonType button, int value){
    assert(floor >= 0);
    assert(floor < N_FLOORS);
    assert(button >= 0);
    assert(button < N_BUTTONS);

    writePort(&serialLine, btnLights[floor][button], value);
}

void elevio_floorIndicator(int floor){
    assert(floor >= 0);
    assert(floor < N_FLOORS);

    if(floor & 0x02){
        writePort(&serialLine, PORT_FLOOR_HI, 1);
    } else {
        writePort(&serialLine, PORT_FLOOR_HI, 0);
    }    

    if(floor & 0x01){
        writePort(&serialLine, PORT_FLOOR_LO, 1);
    } else {
        writePort(&serialLine, PORT_FLOOR_LO, 0);
    }
}

void elevio_doorOpenLamp(int value){
    writePort(&serialLine, PORT_DOOR, value);
}

void elevio_stopLamp(int value){
    writePort(&serialLine, PORT_LIGHT_STOP, value);
}



static const uint8_t btns[N_FLOORS][N_BUTTONS] = {
    {PIN_BTN_UP0, PIN_BTN_DOWN0, PIN_BTN_CAB0},
    {PIN_BTN_UP1, PIN_BTN_DOWN1, PIN_BTN_CAB1},
    {PIN_BTN_UP2, PIN_BTN_DOWN2, PIN_BTN_CAB2},
    {PIN_BTN_UP3, PIN_BTN_DOWN3, PIN_BTN_CAB3},
};
int elevio_callButton(int floor, ButtonType button){
    return readPin(&serialLine, btns[floor][button]);
}

int elevio_floorSensor(void){
    if(readPin(&serialLine, PIN_FLOOR0)){
        return 0;
    } else if(readPin(&serialLine, PIN_FLOOR1)){
        return 1;
    } else if(readPin(&serialLine, PIN_FLOOR2)){
        return 2;
    } else if(readPin(&serialLine, PIN_FLOOR3)){
        return 3;
    } else {
        return -1;
    }
}

int elevio_stopButton(void){
    return readPin(&serialLine, PIN_BTN_STOP);
}

int elevio_obstruction(void){
    return readPin(&serialLine, PIN_OBSTR);
}

