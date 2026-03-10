# Transpiled from English language source
# array_demo.abc - demonstrates homogeneous typed arrays
print("=== Homogeneous Arrays ===")
print("")
# --- Declare a number array ---
scores = [85, 92, 78, 95, 88]
print("Scores:")
print(scores)
print("Count:")
print(len(scores))
print("Sum:")
print(sum(scores))
print("First score:")
print(scores[0])
print("Last score:")
print(scores[-1])
# --- Append (type-checked) ---
scores = scores + [91]
print("After adding 91:")
print(scores)
# --- Element access ---
print("Score at position 2:")
print(scores[2])
# --- Iterate ---
print("")
print("All scores:")
for score in scores:
    print(score)
# --- Text array ---
print("")
print("=== Text Array ===")
names = ["Alice", "Bob", "Carol"]
print(names)
names = names + ["Dave"]
print(len(names))
print(names[0])
print(names[-1])
# --- Type safety ---
print("")
print("=== Type Safety ===")
print("Appending wrong type raises TypeError (demonstration skipped).")
print("Use 'cast to' for explicit conversion if needed.")
