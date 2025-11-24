#!/bin/bash
# Script to fix SQLite version issue by upgrading to Python 3.11

echo "Fixing SQLite version issue..."
echo ""

# Check if Python 3.11 is available
if ! command -v python3.11 &> /dev/null; then
    echo "❌ Python 3.11 not found. Installing via Homebrew..."
    brew install python@3.11
fi

echo "✓ Python 3.11 found"
echo ""

# Backup current venv
if [ -d "venv" ]; then
    echo "Backing up current venv..."
    mv venv venv.backup.$(date +%Y%m%d_%H%M%S)
fi

# Create new venv with Python 3.11
echo "Creating new virtual environment with Python 3.11..."
python3.11 -m venv venv

# Activate and install dependencies
echo "Installing dependencies..."
source venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt

# Verify SQLite version
echo ""
echo "Verifying SQLite version..."
python -c "import sqlite3; version = sqlite3.sqlite_version; print(f'SQLite version: {version}'); exit(0 if version >= '3.35.0' else 1)"

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ SQLite version is compatible!"
    echo "✅ Virtual environment upgraded successfully"
    echo ""
    echo "You can now run: python main.py"
else
    echo ""
    echo "⚠️  SQLite version may still be too old"
    echo "Try installing pysqlite3: pip install pysqlite3"
fi

