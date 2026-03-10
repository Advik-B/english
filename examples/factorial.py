# Transpiled from English language source
# Factorial Example
# Calculates factorial using a recursive function
print("=== Factorial Calculator ===")


# Define a recursive factorial function
def factorial(n):
    if n <= 1:
        return 1
    smaller = 0
    smaller = factorial(n - 1)
    return n * smaller

# Calculate and print factorials for 0 through 7
print("Calculating factorials:")
i = 0
result = 0
while i <= 7:
    result = factorial(i)
    print(i, "! =", result)
    i = i + 1
print("=== Done! ===")
