# Transpiled from English language source
# GCD and LCM Example
# Calculates Greatest Common Divisor and Least Common Multiple
print("=== GCD and LCM Calculator ===")
# Function to calculate GCD using Euclidean algorithm
def gcd(a, b):
    temp = 0
    while b != 0:
        temp = b
        b = a % b
        a = temp
    return a

# Function to calculate LCM using GCD
def lcm(a, b):
    g = 0
    g = gcd(a, b)
    return (a * b) / g

# Test with various pairs
print("Testing with 48 and 18:")
g1 = 0
g1 = gcd(48, 18)
l1 = 0
l1 = lcm(48, 18)
print("GCD(48, 18) =", g1)
print("LCM(48, 18) =", l1)
print("Testing with 12 and 15:")
g2 = 0
g2 = gcd(12, 15)
l2 = 0
l2 = lcm(12, 15)
print("GCD(12, 15) =", g2)
print("LCM(12, 15) =", l2)
print("Testing with 100 and 25:")
g3 = 0
g3 = gcd(100, 25)
l3 = 0
l3 = lcm(100, 25)
print("GCD(100, 25) =", g3)
print("LCM(100, 25) =", l3)
print("=== Done! ===")
