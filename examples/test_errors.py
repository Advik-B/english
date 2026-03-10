# Transpiled from English language source
from typing import Final

# Test file to demonstrate error messages
# This will cause an undefined variable error with suggestion
myVariable = 10
print(myVarible)
# Typo: myVarible instead of myVariable
# This will cause an undefined function error
myFunctoin()
# Typo: myFunctoin
# This will cause a constant reassignment error
PI: Final = 3.14
PI = 3
# Error: cannot reassign constant
# This will cause an argument count mismatch
def add(a, b):
    return a + b

result = add(5)
# Missing second argument
# This will cause a division by zero error
x = 10
y = 0
result = x / y
