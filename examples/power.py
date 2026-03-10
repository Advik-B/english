# Transpiled from English language source
# Power Function Example
# Calculates x^n using iteration
print("=== Power Calculator ===")
# Function to calculate power
def power(base, exponent):
    result = 1
    i = 0
    while i < exponent:
        result = result * base
        i = i + 1
    return result

# Test various powers
p1 = 0
p1 = power(2, 0)
print("2^0 =", p1)
p2 = 0
p2 = power(2, 5)
print("2^5 =", p2)
p3 = 0
p3 = power(3, 4)
print("3^4 =", p3)
p4 = 0
p4 = power(10, 3)
print("10^3 =", p4)
p5 = 0
p5 = power(5, 2)
print("5^2 =", p5)
# Powers of 2 table
print("Powers of 2 table:")
n = 0
val = 0
while n <= 10:
    val = power(2, n)
    print("2^", n, "=", val)
    n = n + 1
print("=== Done! ===")
