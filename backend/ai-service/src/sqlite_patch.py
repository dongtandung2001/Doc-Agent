"""
SQLite version patch for ChromaDB compatibility on macOS.
This patches Python's sqlite3 module to use the system SQLite library.
"""
import sys
import os
import ctypes.util

# On macOS, try to use the system SQLite which is usually newer
if sys.platform == 'darwin':
    try:
        # Try to find system SQLite library
        sqlite_lib_path = ctypes.util.find_library('sqlite3')
        if sqlite_lib_path:
            # Try to use pysqlite3 if available
            try:
                import pysqlite3
                sys.modules['sqlite3'] = pysqlite3
            except ImportError:
                # If pysqlite3 not available, try to load system SQLite
                # This is a workaround - the built-in sqlite3 will still be used
                # but we can at least try
                pass
    except Exception:
        pass

# Alternative: Try to install/use pysqlite3
try:
    import pysqlite3
    sys.modules['sqlite3'] = pysqlite3
except ImportError:
    # pysqlite3 not available - will need to install it or upgrade Python
    pass
