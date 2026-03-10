# Transpiled from English language source
# FizzBuzz Example
# Classic programming challenge: Print numbers 1-20
# - Print "Fizz" for multiples of 3
# - Print "Buzz" for multiples of 5
# - Print "FizzBuzz" for multiples of both 3 and 5
# - Print the number otherwise
print("=== FizzBuzz ===")
print("Numbers 1-20 with FizzBuzz rules:")
i = 1
mod3 = 0
mod5 = 0
while i <= 20:
    mod3 = i % 3
    mod5 = i % 5
    if mod3 == 0:
        if mod5 == 0:
            print("FizzBuzz")
        else:
            print("Fizz")
    elif mod5 == 0:
        print("Buzz")
    else:
        print(i)
    i = i + 1
print("=== Done! ===")
