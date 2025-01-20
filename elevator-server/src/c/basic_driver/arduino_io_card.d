import std;
import core.sync.mutex;
import core.thread;

struct SerialLine {
    FileDescriptor  handle;
    Mutex           mutex;
}

enum Port : ubyte {
    lightUp0    = 30,
    lightUp1    = 31,
    lightUp2    = 29,
    lightUp3    = 0,
    lightDown0  = 0,
    lightDown1  = 28,
    lightDown2  = 26,
    lightDown3  = 27,
    lightCab0   = 34,
    lightCab1   = 35,
    lightCab2   = 32,
    lightCab3   = 33,
    
    lightStop   = 37,
    floorHi     = 23,
    floorLo     = 22,
    door        = 24,
    motorDir    = 36,
    
    respA       = 55,
    respB       = 57,
    respC       = 59,
    
    unused12    = 12,
    unused25    = 25,
    
    zero        = 0,
}

enum PWMPort : ubyte {
    motorSpeed  = 13,
}

enum Pin : ubyte {
    btnUp0      = 46,
    btnUp1      = 47,
    btnUp2      = 45,
    btnUp3      = 0,
    btnDown0    = 0,
    btnDown1    = 44,
    btnDown2    = 42,
    btnDown3    = 43,
    btnCab0     = 50,
    btnCab1     = 51,
    btnCab2     = 48,
    btnCab3     = 49,
        
    btnStop     = 53,
    obstr       = 52,
    floor0      = 40,
    floor1      = 41,
    floor2      = 38,
    floor3      = 39,
    tacho       = 62,
        
    testA       = 54,
    testB       = 56,
    testC       = 58,
}

enum WidePin : ubyte {
    RT          = 54,
    p38         = 38,
    p42         = 42, 
    p46         = 46,
    tacho       = 62,
    ID          = 63,
}

enum WidePinsMask : ubyte {
    testA       = 0x01,
    testB       = 0x04,
    testC       = 0x10,
}




version(Windows){
    import core.sys.windows.winbase;
    import core.sys.windows.winnt;

    alias FileDescriptor = HANDLE;

    private FileDescriptor openSerialHandle(string comPort){
        string port = "\\\\.\\" ~ comPort;
        auto handle = CreateFileA(port.ptr, GENERIC_READ | GENERIC_WRITE, 0, null, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, null);
        assert(handle != INVALID_HANDLE_VALUE, "Unable to open serial line");
        
        DCB serialParams;
        GetCommState(handle, &serialParams);
        serialParams.BaudRate   = 115200;
        serialParams.ByteSize   = 8;    
        serialParams.StopBits   = ONESTOPBIT;
        serialParams.Parity     = NOPARITY;
        SetCommState(handle, &serialParams);
        
        COMMTIMEOUTS serialTimeouts;
        GetCommTimeouts(handle, &serialTimeouts);
        serialTimeouts.ReadIntervalTimeout          = DWORD.max;
        serialTimeouts.ReadTotalTimeoutMultiplier   = DWORD.max;
        serialTimeouts.ReadTotalTimeoutConstant     = 100; // 100 ms to respond after trying to read from the device
        serialTimeouts.WriteTotalTimeoutMultiplier  = 0;
        serialTimeouts.WriteTotalTimeoutConstant    = 0;
        
        SetCommTimeouts(handle, &serialTimeouts);
        
        return handle;
    }
    
    string autoDetectArduino(){
        
        auto comPorts = executeShell(join([
			"powershell -command \"",
				"Get-WmiObject Win32_PnPEntity | ",
				"where { $_.name -match '\\(COM\\d+\\)' } | ",
				"ForEach-Object { ",
                    "try {",
                        "$a = $_.GetDeviceProperties('DEVPKEY_DEVICE_BusReportedDeviceDesc').DeviceProperties.Data + ' : ' ",
                    "}",
                    "catch {}",
                    "$a + $_.name ",
                "}",
			"\""]))
            .output
            .splitLines;
            
        auto ioCards = comPorts
            .filter!(a => a.startsWith("NTNU Arduino IO Card"))
            .map!(a => a.matchFirst(r"\((COM\d+)\)"))
            .array;
        if(ioCards.length){
            return ioCards[0][1];
        }
        
        // GetDeviceProperties() does not exist on Windows 7, so we can't get BusReportedDeviceDesc.
        // Return the first Arduino instead
        auto arduinos = comPorts
            .filter!(a => a.startsWith("Arduino Mega 2560"))
            .map!(a => a.matchFirst(r"\((COM\d+)\)"))
            .array
            .sort!((a, b) => a[1] < b[1]);
        if(arduinos.length){
            return arduinos[0][1];
        }
        
        writeln("Unable to auto detect Arduino!");
        writeln("Serial ports reported by \"Get-WmiObject Win32_PnPEntity\"");
        writefln!("%-(  %s\n%)")(comPorts);
        return "";
    }
}

version(Posix){
    import core.sys.posix.termios;
    import core.sys.posix.fcntl;
    import core.sys.posix.unistd;
    import core.sys.posix.sys.ioctl;
    import core.sys.linux.termios;
    import core.stdc.errno;
    
    alias FileDescriptor = int;
    
    private FileDescriptor openSerialHandle(string comPort){
        
        import std.string : toStringz;
        int r;
        
        auto handle = open(comPort.toStringz, (O_RDWR | O_NOCTTY));
        assert(handle >= 0, format!("Unable to open serial line (errno:%d)")(errno));
        

        termios tty;
        r = tcgetattr(handle, &tty);
        assert(r >= 0);
        
        tty.c_cflag = (B115200 | CLOCAL | CREAD | CS8);
        tty.c_lflag = 0;
        tty.c_iflag = 0;
        tty.c_oflag = 0;
        tty.c_cc[VMIN] = 0;
        tty.c_cc[VTIME] = 10;   // 10 tenths of a second to respond after trying to read from the device
        
        
        tcflush(handle, TCIFLUSH);
        r = tcsetattr(handle, TCSANOW, &tty);
        assert(r >= 0);

        return handle;
    }
    
    string autoDetectArduino(){
    
        auto exec = executeShell("readlink -f $(find /dev/serial/by-id/ -name '*NTNU_Arduino_IO_Card*' | head -n 1)");
        if(!exec.status){
            return exec.output.strip;
        }
    
        writeln("Unable to auto detect Arduino!");
    
        string[] ttys = executeShell("dmesg | grep tty")
            .output
            .splitLines
            .map!(a => a.matchFirst(r".+(tty[a-zA-Z0-9]+).+")[1])
            .uniq
            .array;
            
        writeln("TTYs reported by \"dmesg | grep tty\":");
        foreach(tty; ttys){
            writefln!("  %s")(tty);
        }
        return "";
    }
}

SerialLine openSerial(string port, bool reentrant = true){

    SerialLine s = {
        handle: openSerialHandle(port),
        mutex:  reentrant ? new Mutex() : null,
    };
    
    
    scope(failure) writeln("Serial line is not responding as an NTNU Arduino IO Card"); 
    ubyte v;
    foreach(_; 0..10){
        try {
            v = s.readWidePin(WidePin.ID);
        } catch(Throwable t){}
        if(v == 0xa2){
            break;
        }
        writeln("Retrying...");
    }
    assert(v == 0xa2, format!("Incorrect response (got 0x%02x instead of 0xa2)")(v));
    
    return s;
}

private void writeByteImpl(FileDescriptor handle, ubyte data){
    scope(failure){
        writeln("Unable to write to serial line.");
    }
    version(Windows){
        uint n;
        auto r = WriteFile(handle, &data, 1, &n, null);
        assert(r);
    }
    version(Posix){
        auto r = core.sys.posix.unistd.write(handle, &data, 1);
        assert(r == 1);
    }
}

private ubyte readByteImpl(FileDescriptor handle){
    scope(failure){
        writeln("No data from serial line before timing out. Device did not respond to request-to-read");
    }
    ubyte buf;
    version(Windows){
        uint n;        
        auto r = ReadFile(handle, &buf, 1, &n, null);
        assert(r && n);
    }
    version(Posix){
        auto r = core.sys.posix.unistd.read(handle, &buf, 1);
        assert(r);
    }
    return buf;
}


void writePort(SerialLine serial, Port port, bool v){
    if(v){
        port |= 0b01000000;
    }
    
    if(serial.mutex) serial.mutex.lock;
    scope(exit) if(serial.mutex) serial.mutex.unlock;
    
    writeByteImpl(serial.handle, port);
}

void writePWM(SerialLine serial, PWMPort port, ubyte width){
    if(serial.mutex) serial.mutex.lock;
    scope(exit) if(serial.mutex) serial.mutex.unlock;
    
    writeByteImpl(serial.handle, port);
    writeByteImpl(serial.handle, width);
}

bool readPin(SerialLine serial, Pin pin){
    scope(failure){
        writefln("Error reading pin 0x%02x / %d", pin, pin);
    }
    ubyte msg = pin | 0b10000000;
    ubyte ret = 0;
    
    if(serial.mutex) serial.mutex.lock;
    scope(exit) if(serial.mutex) serial.mutex.unlock;
    
    writeByteImpl(serial.handle, msg);
    ret = readByteImpl(serial.handle) > 0;
    
    return ret != 0;
}

ubyte readWidePin(SerialLine serial, WidePin startingPin){
    scope(failure){
        writefln("Error reading wide pins 0x%02x / %d", startingPin, startingPin);
    }
    ubyte msg = startingPin | 0b11000000;
    ubyte ret = 0;
    
    if(serial.mutex) serial.mutex.lock;
    scope(exit) if(serial.mutex) serial.mutex.unlock;
    
    writeByteImpl(serial.handle, msg);
    ret = readByteImpl(serial.handle);
    
    return ret;
}



