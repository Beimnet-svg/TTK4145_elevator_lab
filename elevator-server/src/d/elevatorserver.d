// Run with:        rdmd -I../basic_driver elevatorserver.d 
// Compile with:    dmd ../basic_driver/arduino_io_card.d elevatorserver.d -ofelevatorserver

import std;
import core.thread;

import arduino_io_card;


struct PinValues {
    union {
        ubyte PIN38;
        mixin(bitfields!(
            bool, "up2",    1,
            bool, "down1",  1,
            bool, "down3",  1,
            bool, "down2",  1,
            bool, "floor1", 1,
            bool, "floor0", 1,
            bool, "floor3", 1,
            bool, "floor2", 1,
        ));
    }
    union {
        ubyte PIN46;
        mixin(bitfields!(
            bool, "stop",   1,
            bool, "obstr",  1,
            bool, "cab1",   1,
            bool, "cab0",   1,
            bool, "cab3",   1,
            bool, "cab2",   1,
            bool, "up1",    1,
            bool, "up0",    1,
        ));
    }
    union {
        ubyte PIN62;
        ubyte tacho;
    }
}


immutable Port[4][3] lights = [
    [Port.lightUp0,     Port.lightUp1,      Port.lightUp2,      Port.lightUp3   ],
    [Port.lightDown0,   Port.lightDown1,    Port.lightDown2,    Port.lightDown3 ],
    [Port.lightCab0,    Port.lightCab1,     Port.lightCab2,     Port.lightCab3  ],
];
void writeButtonLight(SerialLine serial, ubyte floor, ubyte btn, bool value){
    serial.writePort(lights[floor][btn], value);
}

void writeDoorLight(SerialLine serial, bool value){
    serial.writePort(Port.door, value);
}

void writeStopLight(SerialLine serial, bool value){
    serial.writePort(Port.lightStop, value);
}

void writeFloorIndicator(SerialLine serial, ubyte floor){
    // Binary encoding. One light must always be on.
    if (floor & 0x02) {
        serial.writePort(Port.floorHi, true);
    } else {
        serial.writePort(Port.floorHi, false);
    }    

    if (floor & 0x01) {
        serial.writePort(Port.floorLo, true);
    } else {
        serial.writePort(Port.floorLo, false);
    }
}

void writeMotorSpeed(SerialLine serial, ubyte dir, ubyte speed){
    if(dir == 0){
        serial.writePWM(PWMPort.motorSpeed, 0);
    } else if(dir < 128){
        serial.writePort(Port.motorDir, false);
        serial.writePWM(PWMPort.motorSpeed, speed);
    } else {
        serial.writePort(Port.motorDir, true);
        serial.writePWM(PWMPort.motorSpeed, speed);
    }
}




bool readButton(PinValues pinValues, ubyte floor, ubyte btn){
    bool[4][3] buttons = [
        [pinValues.up0,     pinValues.up1,      pinValues.up2,      false],
        [false,             pinValues.down1,    pinValues.down2,    pinValues.down3],
        [pinValues.cab0,    pinValues.cab1,     pinValues.cab2,     pinValues.cab3],
    ];
    return buttons[floor][btn];
}

bool readStopButton(PinValues pinValues){
    return pinValues.stop;
}

bool readObstruction(PinValues pinValues){
    return pinValues.obstr;
}

ubyte readTacho(PinValues pinValues){
    return pinValues.tacho;
}

ubyte readFloor(PinValues pinValues){
    if(pinValues.floor0){
        return 0;
    } else if (pinValues.floor1){
        return 1;
    } else if (pinValues.floor2){
        return 2;
    } else if (pinValues.floor3){
        return 3;
    } else {
        return ubyte.max;
    }
}

byte readMotorSpeed(PinValues pinValues){
    return pinValues.tacho;
}



PinValues readPinValues(SerialLine serial){
    PinValues pinValues;
    pinValues.PIN38 = serial.readWidePin(WidePin.p38);
    pinValues.PIN46 = serial.readWidePin(WidePin.p46);
    pinValues.PIN62 = serial.readWidePin(WidePin.tacho);
    //writefln!("PinValues(%08b, %08b, %3d)")(pinValues.PIN38, pinValues.PIN46, pinValues.tacho);
    return pinValues;
}



void main(string[] args){

    ushort      tcpPort             = 15657;
    string      serialPort          = autoDetectArduino();
    auto        pinPollInterval_ms  = 10;
    
    if(serialPort == ""){
        writeln("Arduino IO Card not detected - supply a serial port on the command line with '-s serialPort'");
        return;
    }
    
    getopt(args,
        "p|port",       &tcpPort,
        "s|serial",     &serialPort,
        "i|interval",   &pinPollInterval_ms,
    );

    writefln!("TCP port:    %d")(tcpPort);
    writefln!("Serial port: %s")(serialPort);
    writefln!("Poll rate:   %s ms")(pinPollInterval_ms);
    
    Duration    pinPollInterval     = pinPollInterval_ms.msecs;
    SerialLine  serial              = openSerial(serialPort, true);
    PinValues   pinValues           = serial.readPinValues;
    MonoTime    lastReadTime        = MonoTime.currTime;
    Socket      acceptSock          = new TcpSocket();
    ubyte[4]    buf;
    
    serial.writeMotorSpeed(0, 0);
    
    acceptSock.setOption(SocketOptionLevel.SOCKET, SocketOption.REUSEADDR, 1);
    acceptSock.bind(new InternetAddress(tcpPort));
    acceptSock.listen(1);    
    
    
    writefln!("Elevator server started");
    scope(exit) writeln("[", cast(DateTime)Clock.currTime, "]: Shut down");
    
    while(true){
        auto driverSock = acceptSock.accept();
        writeln("[", cast(DateTime)Clock.currTime, "]: Connected to ", driverSock.remoteAddress);
        while(driverSock.isAlive){
            buf = 0;
            auto n = driverSock.receive(buf);

            if(n <= 0){
                serial.writeMotorSpeed(0, 0);
                driverSock.close();
                writeln("[", cast(DateTime)Clock.currTime, "]: Disconnected");
            } else {
                MonoTime now = MonoTime.currTime;
                if(now - lastReadTime > pinPollInterval){
                    pinValues = serial.readPinValues;
                    lastReadTime = now;
                }
                
                switch(buf[0]){
                case 1:
                    serial.writeMotorSpeed(buf[1], buf[2] ? buf[3] : 255);
                    break;
                case 2:
                    serial.writeButtonLight(buf[1], buf[2], buf[3]>0);
                    break;
                case 3:
                    serial.writeFloorIndicator(buf[1]);
                    break;
                case 4:
                    serial.writeDoorLight(buf[1]>0);
                    break;
                case 5:
                    serial.writeStopLight(buf[1]>0);
                    break;
                    
                case 6:
                    buf[1..$] = [pinValues.readButton(buf[1], buf[2]).to!ubyte, 0, 0];
                    driverSock.send(buf);
                    break;
                case 7:
                    auto v = pinValues.readFloor();
                    buf[1..$] = (v == v.max) ? [0, 0, 0] : [1, v.to!ubyte, 0];
                    driverSock.send(buf);
                    break;
                case 8:
                    buf[1..$] = [pinValues.readStopButton().to!ubyte, 0, 0];
                    driverSock.send(buf);
                    break;
                case 9:
                    buf[1..$] = [pinValues.readObstruction().to!ubyte, 0, 0];
                    driverSock.send(buf);
                    break;
                case 10:
                    buf[1..$] = [pinValues.readTacho(), 0, 0];
                    driverSock.send(buf);
                    break;                
                    
                default:
                    break;
                }
            }
        }
    }
}





