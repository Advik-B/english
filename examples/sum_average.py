# Transpiled from English language source
# Sum and Average Example
# Calculates sum and average of numbers
print("=== Sum and Average Calculator ===")
numbers = [10, 20, 30, 40, 50, 60, 70, 80, 90, 100]
print("Numbers: 10, 20, 30, 40, 50, 60, 70, 80, 90, 100")
# Calculate sum
sum = 0
for item in numbers:
    sum = sum + item
print("Sum:", sum)
# Calculate average
count = 10
average = sum / count
print("Count:", count)
print("Average:", average)
# Find minimum and maximum
minimum = numbers[0]
maximum = numbers[0]
for item in numbers:
    if item < minimum:
        minimum = item
    if item > maximum:
        maximum = item
print("Minimum:", minimum)
print("Maximum:", maximum)
print("=== Done! ===")
