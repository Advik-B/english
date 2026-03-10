# Transpiled from English language source
# Pattern Printing Example
# Prints various patterns using loops
print("=== Pattern Examples ===")
# Triangle pattern using numbers
print("Number triangle:")
row = 1
col = 1
while row <= 5:
    col = 1
    while col <= row:
        print(col)
        col = col + 1
    row = row + 1
# Countdown pattern
print("Countdown rows:")
i = 5
j = 0
while i >= 1:
    j = i
    while j >= 1:
        print(j)
        j = j - 1
    i = i - 1
# Square pattern with border check
print("Square pattern 4x4:")
r = 1
c = 1
while r <= 4:
    c = 1
    while c <= 4:
        if r == 1:
            print("*")
        elif r == 4:
            print("*")
        elif c == 1:
            print("*")
        elif c == 4:
            print("*")
        else:
            print(" ")
        c = c + 1
    r = r + 1
print("=== Done! ===")
