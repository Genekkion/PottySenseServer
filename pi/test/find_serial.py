import subprocess

def find_serial():
    # Running the command and capturing the output
    result = subprocess.run(['python3.11', '-m', 'serial.tools.list_ports'], capture_output=True, text=True)
    
    # Splitting the output into lines
    output_lines = result.stdout.splitlines()
    
    # Checking if there's at least one line of output to print
    if output_lines:
        first_line = output_lines[0]
        # print(f"First Serial Port: {first_line}")
        return first_line
    else:
        # print("No serial ports found.")
        return None


