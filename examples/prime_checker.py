# Transpiled from English language source
# Prime Number Checker Example
# Checks if numbers are prime and lists first few primes
print("=== Prime Number Checker ===")


# Function to check if a number is prime
def isPrime(n):
    if n < 2:
        return 0
    if n == 2:
        return 1
    divisor = 2
    rem = 0
    while (divisor * divisor) <= n:
        rem = n % divisor
        if rem == 0:
            return 0
        divisor = divisor + 1
    return 1

# Test individual numbers
print("Testing individual numbers:")
num = 2
result = 0
while num <= 20:
    result = isPrime(num)
    if result == 1:
        print(num, "is prime")
    else:
        print(num, "is not prime")
    num = num + 1
# Count primes up to 50
print("Counting primes up to 50:")
count = 0
i = 2
check = 0
while i <= 50:
    check = isPrime(i)
    if check == 1:
        count = count + 1
    i = i + 1
print("Total prime numbers found:", count)
print("=== Done! ===")
