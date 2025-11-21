#!/usr/bin/env python3
"""Test OpenAI API key validity."""
import os
from dotenv import load_dotenv
from openai import OpenAI

load_dotenv()

api_key = os.getenv('OPENAI_API_KEY')

if not api_key:
    print("✗ OPENAI_API_KEY not found in environment")
    exit(1)

print(f"API Key found: {api_key[:10]}...{api_key[-4:]}")
print(f"Key length: {len(api_key)}")

client = OpenAI(api_key=api_key)

print("\nTesting API key...")
try:
    # Try a simple API call
    models = client.models.list()
    print("✓ API key is valid - can list models")
    
    # Try to get account info
    try:
        # Try a simple chat completion to test quota
        response = client.chat.completions.create(
            model="gpt-3.5-turbo",
            messages=[{"role": "user", "content": "Say 'test'"}],
            max_tokens=5
        )
        print("✓ API key works - can make chat completions")
        print(f"  Response: {response.choices[0].message.content}")
    except Exception as e:
        print(f"✗ Chat completion error: {e}")
        if "quota" in str(e).lower() or "429" in str(e):
            print("  This appears to be a quota/billing issue")
        elif "401" in str(e) or "invalid" in str(e).lower():
            print("  This appears to be an authentication issue")
        
except Exception as e:
    print(f"✗ API key error: {e}")
    if "401" in str(e) or "invalid" in str(e).lower():
        print("  The API key appears to be invalid or expired")
    elif "quota" in str(e).lower() or "429" in str(e):
        print("  This appears to be a quota/billing issue")

