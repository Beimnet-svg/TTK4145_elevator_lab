# Elevator server
In the [TTK4145 elevator project](https://github.com/TTK4145/Project) the elevator hardware is controlled by a server, which exposes a TCP interface. This repository contains the source code for that server, as well as a download link to the executable.

If you are on the lab, you should be able to start the elevator server from any terminal window by just running `elevatorserver`. Likewise, [the simulator](https://github.com/TTK4145/Simulator-v2) should also be runnable with `simelevatorserver`.

## Executables

[The executables for Windows and Linux can be found here.](https://github.com/TTK4145/elevator-server/releases/latest)
 
The server should run in its own terminal, and should not need to be restarted if the client is restarted.

Remember to `chmod +x elevatorserver` in order to give yourself permission to run the downloaded file.

### Running

Once the server has started, it will start listening on `localhost:15657`. You can then connect to it by using a [client](https://github.com/TTK4145?q=driver) that adheres to [the protocol](https://github.com/TTK4145/elevator-server#protocol).


### Building from source

The elevator server is written in D, so you will need [a D compiler](http://dlang.org/download.html#dmd).

Compile with `dmd elevatorserver.d arduino_io_card.d`

---

A low-level C driver is also included for completeness. This is not an elevator server, as it exposes its functionality through C function calls instead of a TCP connection. This means it does not expose an interface similar to the simulator, and therefore its usage is not recommended. Prefer using [the C client](https://github.com/TTK4145/driver-c) with the elevator server instead.

### Protocol

 - All TCP messages must have a length of 4 bytes
 - The instructions for reading from the hardware send replies that are 4 bytes long, where the last byte is always 0
 - The instructions for writing to the hardware do not send any replies
 
 
<table>
    <tbody>
        <tr>
            <td><strong>Writing</strong></td>
            <td align="center" colspan="4">Instruction</td>
        </tr>
        <tr>
            <td><em>Reload config (file and args)</em></td>
            <td>&nbsp;&nbsp;0&nbsp;&nbsp;</td>
            <td>X</td>
            <td>X</td>
            <td>X</td>
        </tr>
        <tr>
            <td><em>Motor direction</em></td>
            <td>&nbsp;&nbsp;1&nbsp;&nbsp;</td>
            <td>direction<br>[-1 (<em>255</em>),0,1]</td>
            <td>custom speed<br>[0,1]</td>
            <td>speed<br>[0..255]<br>(defaults to 255 if <code>custom speed</code> is 0)</td>
        </tr>
        <tr>
            <td><em>Order button light</em></td>
            <td>&nbsp;&nbsp;2&nbsp;&nbsp;</td>
            <td>button<br>[0,1,2]</td>
            <td>floor<br>[0..NF]</td>
            <td>value<br>[0,1]</td>
        </tr>
        <tr>
            <td><em>Floor indicator</em></td>
            <td>&nbsp;&nbsp;3&nbsp;&nbsp;</td>
            <td>floor<br>[0..NF]</td>
            <td>X</td>
            <td>X</td>
        </tr>
        <tr>
            <td><em>Door open light</em></td>
            <td>&nbsp;&nbsp;4&nbsp;&nbsp;</td>
            <td>value<br>[0,1]</td>
            <td>X</td>
            <td>X</td>
        </tr>
        <tr>
            <td><em>Stop button light</em></td>
            <td>&nbsp;&nbsp;5&nbsp;&nbsp;</td>
            <td>value<br>[0,1]</td>
            <td>X</td>
            <td>X</td>
        </tr>
    </tbody>
</table>
<table>
    <tbody>        
        <tr>
            <td><strong>Reading</strong></td>
            <td align="center" colspan="4">Instruction</td>
            <td></td>
            <td align="center" colspan="4">Output</td>
        </tr>
        <tr>
            <td><em>Order button</em></td>
            <td>&nbsp;&nbsp;6&nbsp;&nbsp;</td>
            <td>button<br>[0,1,2]</td>
            <td>floor<br>[0..NF]</td>
            <td>X</td>
            <td align="right"><em>Returns:</em></td>
            <td>6</td>
            <td>pressed<br>[0,1]</td>
            <td>0</td>
            <td>0</td>
        </tr>
        <tr>
            <td><em>Floor sensor</em></td>
            <td>&nbsp;&nbsp;7&nbsp;&nbsp;</td>
            <td>X</td>
            <td>X</td>
            <td>X</td>
            <td align="right"><em>Returns:</em></td>
            <td>7</td>
            <td>at floor<br>[0,1]</td>
            <td>floor<br>[0..NF]</td>
            <td>0</td>
        </tr>
        <tr>
            <td><em>Stop button</em></td>
            <td>&nbsp;&nbsp;8&nbsp;&nbsp;</td>
            <td>X</td>
            <td>X</td>
            <td>X</td>
            <td align="right"><em>Returns:</em></td>
            <td>8</td>
            <td>pressed<br>[0,1]</td>
            <td>0</td>
            <td>0</td>
        </tr>
        <tr>
            <td><em>Obstruction switch</em></td>
            <td>&nbsp;&nbsp;9&nbsp;&nbsp;</td>
            <td>X</td>
            <td>X</td>
            <td>X</td>
            <td align="right"><em>Returns:</em></td>
            <td>9</td>
            <td>active<br>[0,1]</td>
            <td>0</td>
            <td>0</td>
        </tr>
        <tr>
            <td colspan="0"><em>NF = Num floors. X = Don't care.</em></td>
        </tr>
    </tbody>
</table>
 
Button types (for reading the button and setting the button light) are in the order `0: Hall Up`, `1: Hall Down`, `2: Cab`.