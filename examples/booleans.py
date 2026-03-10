# Transpiled from English language source
# Boolean Operations Example
# Demonstrates boolean values and operations
print("=== Boolean Operations ===")
# Boolean declarations
isRaining = True
isSunny = False
print("isRaining:", isRaining)
print("isSunny:", isSunny)
# Boolean in conditionals
print("Checking boolean conditions:")
if isRaining == True:
    print("Bring an umbrella!")
else:
    print("No umbrella needed.")
if isSunny == False:
    print("It's cloudy.")
# Toggle operation
print("Before toggle - isRaining:", isRaining)
isRaining = not isRaining
print("After toggle - isRaining:", isRaining)
isRaining = not isRaining
print("After second toggle - isRaining:", isRaining)
# Boolean in array
flags = [True, False, True, True, False]
print("Boolean array:")
for item in flags:
    print(item)
# Count true values
trueCount = 0
for item in flags:
    if item == True:
        trueCount = trueCount + 1
print("Number of true values:", trueCount)
print("=== Done! ===")
