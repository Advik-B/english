# Transpiled from English language source
# Logical Operators Example
# Demonstrates 'and', 'or', and 'not' operators
print("=== Logical Operators ===")
age = 25
has_license = True
is_tired = False
# AND operator: both conditions must be true
print("Testing AND:")
if (age >= 18) and (has_license == True):
    print("Can drive: age OK and has license.")
# OR operator: at least one condition must be true
print("Testing OR:")
if (age < 16) or (is_tired == True):
    print("Should not drive.")
else:
    print("OK to drive (neither too young nor tired).")
# NOT operator: inverts a boolean
print("Testing NOT:")
if not is_tired:
    print("Driver is not tired - good!")
# Chaining multiple operators
print("Testing chained operators:")
speed = 60
speed_limit = 70
road_wet = True
if (speed < speed_limit) and not road_wet:
    print("Driving safely on dry road.")
elif (speed < speed_limit) and road_wet:
    print("Driving safely but road is wet - be careful!")
else:
    print("Speeding! Slow down.")
# Complex logical conditions
print("FizzBuzz with logical operators:")
n = 1
while n <= 15:
    by3 = n % 3
    by5 = n % 5
    if (by3 == 0) and (by5 == 0):
        print("FizzBuzz")
    elif by3 == 0:
        print("Fizz")
    elif by5 == 0:
        print("Buzz")
    else:
        print(n)
    n = n + 1
print("=== Done! ===")
