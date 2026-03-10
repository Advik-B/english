# Transpiled from English language source
# Fibonacci Sequence Example
# Calculates and prints the first 10 Fibonacci numbers
print("=== Fibonacci Sequence ===")
print("Calculating the first 10 Fibonacci numbers:")
a = 0
b = 1
temp = 0
count = 0
while count < 10:
    print(a)
    temp = a + b
    a = b
    b = temp
    count = count + 1
print("=== Done! ===")
