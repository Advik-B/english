# Transpiled from English language source
def _table_remove(d, k):
    result = dict(d)
    result.pop(k, None)
    return result

# lookup_table_demo.abc - demonstrates lookup tables (ordered key-value dictionaries)
print("=== Lookup Table Demo ===")
print("")
# --- Create and populate ---
ages = {}
ages["Alice"] = 30
ages["Bob"] = 25
ages["Carol"] = 35
print("Age of Alice:")
print(ages["Alice"])
print("Age of Bob:")
print(ages["Bob"])
print("Number of entries:")
print(len(ages))
# --- Alternative access syntax ---
print("")
print("=== Entry syntax ===")
print("Carol's age via 'the entry' syntax:")
print(ages["Carol"])
# --- Key membership test (has) ---
print("")
print("=== Has Key ===")
if "Alice" in ages:
    print("Table has Alice.")
if "Dave" in ages:
    print("Table has Dave.")
else:
    print("Table does not have Dave.")
# --- Iterate over keys ---
print("")
print("=== All entries (insertion order) ===")
for name in ages:
    print(name)
# --- keys() and values() stdlib functions ---
print("")
print("=== keys() and values() ===")
print(list(ages.keys()))
print(list(ages.values()))
# --- Remove an entry ---
print("")
print("=== After removing Bob ===")
ages = _table_remove(ages, "Bob")
print(len(ages))
if "Bob" in ages:
    print("Bob still present.")
else:
    print("Bob removed successfully.")
# --- Number keys ---
print("")
print("=== Number Keys ===")
squares = {}
squares[1] = 1
squares[2] = 4
squares[3] = 9
squares[4] = 16
print("Square of 3:")
print(squares[3])
if 2 in squares:
    print("Has key 2.")
# --- Boolean keys ---
print("")
print("=== Boolean Keys ===")
flags = {}
flags[True] = "yes"
flags[False] = "no"
print(flags[True])
print(flags[False])
