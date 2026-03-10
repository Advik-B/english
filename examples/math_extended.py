# Transpiled from English language source
import math
import random

# Math Constants and Functions Example
# Demonstrates pi, e, infinity, log, exp, random, and more
print("=== Math Constants ===")
print("Pi:")
print(math.pi)
print("Euler's number (e):")
print(math.e)
print("Infinity:")
print(math.inf)
print("")
print("=== Math Functions ===")
# Logarithms
print("Natural log of 100:")
print(math.log(100))
print("Log base 10 of 1000:")
print(math.log10(1000))
print("Log base 2 of 1024:")
print(math.log2(1024))
# Exponential
print("e to the power of 2 (exp(2)):")
print(math.exp(2))
# Verify: exp(1) should equal e
result = math.exp(1)
print("exp(1) equals e:")
print(result)
# Circle area using pi
radius = 5
area = (math.pi * radius) * radius
print("Area of circle with radius 5:")
print(area)
# Random numbers
print("")
print("=== Random Numbers ===")
print("Three random numbers between 0 and 1:")
print(random.random())
print(random.random())
print(random.random())
print("Three random integers between 1 and 100:")
print(math.floor(random.uniform(1, 101)))
print(math.floor(random.uniform(1, 101)))
print(math.floor(random.uniform(1, 101)))
print("=== Done! ===")
