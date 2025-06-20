#!/usr/bin/env python3
"""
Script naming convention validator
Validates that script files follow the naming convention: {interval}_{interpreter}_{name}.{ext}
"""

import os
import re
import sys
from pathlib import Path

def validate_script_name(filename):
    """Validate if a script filename follows the naming convention"""
    
    # Remove file extension
    name_without_ext = os.path.splitext(filename)[0]
    
    # Split by underscore
    parts = name_without_ext.split('_')
    
    if len(parts) < 2:
        return False, f"Filename must have at least 2 parts separated by underscore: {filename}"
    
    # Check interval (first part)
    interval = parts[0]
    interval_pattern = r'^(\d+)(ns|us|µs|ms|s|m|h)$'
    if not re.match(interval_pattern, interval):
        return False, f"Invalid interval format: {interval}. Must be like '5s', '10s', '1m', etc."
    
    # Check interpreter (second part)
    interpreter = parts[1]
    valid_interpreters = ['python', 'python3', 'sh', 'bash']
    if interpreter not in valid_interpreters:
        return False, f"Invalid interpreter: {interpreter}. Must be one of {valid_interpreters}"
    
    # Check file extension
    ext = os.path.splitext(filename)[1]
    if interpreter in ['python', 'python3'] and ext != '.py':
        return False, f"Python interpreter {interpreter} must have .py extension"
    elif interpreter in ['sh', 'bash'] and ext != '.sh':
        return False, f"Shell interpreter {interpreter} must have .sh extension"
    
    return True, "Valid"

def main():
    """Main validation function"""
    script_dir = Path(__file__).parent
    all_valid = True
    
    print("Validating script naming convention...")
    print("=" * 50)
    
    # Get all script files
    script_files = []
    for ext in ['.py', '.sh']:
        script_files.extend(script_dir.glob(f"*{ext}"))
        script_files.extend(script_dir.glob(f"py/*{ext}"))
    
    for script_file in script_files:
        if script_file.name == 'validate_naming.py':
            continue  # Skip this validation script
            
        is_valid, message = validate_script_name(script_file.name)
        status = "✅" if is_valid else "❌"
        print(f"{status} {script_file.name}: {message}")
        
        if not is_valid:
            all_valid = False
    
    print("=" * 50)
    if all_valid:
        print("✅ All script files follow the naming convention!")
        return 0
    else:
        print("❌ Some script files do not follow the naming convention.")
        return 1

if __name__ == "__main__":
    sys.exit(main()) 