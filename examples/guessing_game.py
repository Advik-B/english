# Transpiled from English language source
# Guessing Game Simulation
# Simulates a number guessing game
print("=== Number Guessing Simulation ===")
# The secret number
secret = 42
guesses = [20, 50, 35, 45, 40, 42]
attempts = 0
print("Secret number is", secret)
print("Simulating guesses: 20, 50, 35, 45, 40, 42")
for item in guesses:
    attempts = attempts + 1
    print("Guess #", attempts, ":", item)
    if item == secret:
        print("Correct! You found the number!")
        break
    elif item < secret:
        print("Too low!")
    else:
        print("Too high!")
print("Game over! Total attempts:", attempts)
print("=== Done! ===")
