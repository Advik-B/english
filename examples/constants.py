# Transpiled from English language source
from typing import Final

# Constants Example
# Demonstrates constant (immutable) variables
print("=== Constants Example ===")
# Declare constants using "always"
PI: Final = 3.14159
E: Final = 2.71828
MAX_SIZE: Final = 100
APP_NAME: Final = "English Calculator"
print("Mathematical Constants:")
print("PI =", PI)
print("E =", E)
print("Application Constants:")
print("MAX_SIZE =", MAX_SIZE)
print("APP_NAME =", APP_NAME)
# Using constants in calculations
print("Circle calculations using PI:")
radius = 5
print("For radius =", radius)
circumference = (2 * PI) * radius
print("Circumference =", circumference)
area = (PI * radius) * radius
print("Area =", area)
# Regular variable can be changed
counter = 0
counter = counter + 1
print("Counter (mutable):", counter)
# Note: Attempting to change a constant would cause a runtime error
# Set PI to be 3.  # This would fail!
print("=== Done! ===")
