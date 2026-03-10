# Transpiled from English language source
# Continue Statement Example
# Demonstrates skipping loop iterations with 'Continue.'
print("=== Continue Statement ===")
# Print only odd numbers using continue
print("Odd numbers from 1 to 10:")
i = 1
while i <= 10:
    mod = i % 2
    i = i + 1
    if mod == 0:
        continue
    print(i - 1)
print("")
# Skip multiples of 3 in a for-each loop
print("Numbers 1-10 skipping multiples of 3:")
numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
for n in numbers:
    m = n % 3
    if m == 0:
        continue
    print(n)
print("=== Done! ===")
