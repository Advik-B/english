# Transpiled from English language source
# Loops Example
# Demonstrates various loop constructs
print("=== Loop Examples ===")
# Simple repeat N times loop
print("Repeating 3 times:")
for _ in range(3):
    print("Hello!")
# While loop
print("While loop (count from 1 to 5):")
counter = 1
while counter <= 5:
    print("Count:", counter)
    counter = counter + 1
# For-each loop with array
print("For-each loop with array:")
fruits = ["Apple", "Banana", "Cherry"]
for item in fruits:
    print("Fruit:", item)
# Forever loop with break
print("Forever loop with break at 3:")
n = 0
while True:
    n = n + 1
    print("n =", n)
    if n == 3:
        break
print("Exited the forever loop")
# Nested loops
print("Nested loops (2x3 grid):")
row = 1
col = 1
while row <= 2:
    col = 1
    while col <= 3:
        print("Row", row, "Col", col)
        col = col + 1
    row = row + 1
print("=== Done! ===")
