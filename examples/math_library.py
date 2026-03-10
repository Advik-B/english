# Transpiled from English language source
from typing import Final

# Math Library
# A collection of useful math functions
# Function to calculate the square of a number
def square(x):
    return x * x


# Function to calculate the cube of a number
def cube(x):
    return (x * x) * x


# Function to check if a number is even
def isEven(n):
    rem = n % 2
    if rem == 0:
        return True
    else:
        return False

# A useful constant
pi: Final = 3.14159
