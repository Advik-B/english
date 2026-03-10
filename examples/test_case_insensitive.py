# Transpiled from English language source
from typing import Final

# Test case insensitivity
x = 5
y: Final = 10
print("Testing case insensitivity...")
print(x)
print(y)
x = 20
print("x is now:")
print(x)


# Mixed case function
def greet(name):
    print("Hello, ")
    print(name)

result = greet("World")
# Test loops
for _ in range(3):
    print("Loop iteration")
if x == 20:
    print("x equals 20!")
else:
    print("x is not 20")
