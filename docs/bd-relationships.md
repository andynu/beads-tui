# Beads Dependency Relationships: Empirical Findings

This document records the behavior of beads (`bd`) dependency relationships, specifically how they affect the `bd ready` command. These findings inform how the beads-tui should display and interpret dependencies.

## Test Environment

```bash
mkdir -p /tmp/beads-dep-test && cd /tmp/beads-dep-test && bd init --quiet --prefix test
```

Created test issues:
- `test-ac0`: Epic: Parent task (P2, epic)
- `test-3q9`: Child task A (P2, task)
- `test-c7u`: Child task B (P2, task)
- `test-n9r`: Blocker task (P1, task)

---

## Finding 1: "blocks" Dependency Semantics

**Command:** `bd dep add A B --type blocks`

**Meaning:** "A depends on B" / "A is blocked by B" / "B blocks A"

**Example:**
```bash
bd dep add test-3q9 test-n9r --type blocks
# Output: ✓ Added dependency: test-3q9 depends on test-n9r (blocks)
```

**Result:** `test-3q9` (Child A) is now blocked. `bd ready` no longer shows it.

**Display in `bd show`:**
- On the **blocked issue** (test-3q9): `Dependencies: [blocks] test-n9r`
- On the **blocker issue** (test-n9r): `Dependents: [blocks] test-3q9`

**Key insight:** The word "blocks" is used in both places, which can be confusing. The relationship is stored on the blocked issue, pointing to its blocker.

---

## Finding 2: Parent-Child Relationships Don't Inherently Block

**Command:** `bd dep add CHILD PARENT --type parent-child`

**Meaning:** "CHILD is a child of PARENT"

**Example:**
```bash
bd dep add test-3q9 test-ac0 --type parent-child  # Child A is child of Epic
bd dep add test-c7u test-ac0 --type parent-child  # Child B is child of Epic
```

**Result:** Both children remain ready. Parent-child is organizational, not blocking.

```
📋 Ready work (3 issues with no blockers):
1. [P1] test-n9r: Blocker task
2. [P2] test-ac0: Epic: Parent task
3. [P2] test-c7u: Child task B
```

(Child A was already blocked by Blocker via a separate blocks dependency)

---

## Finding 3: Blocking Propagates Through Parent-Child via Blocks

**Scenario:** Epic is blocked by Blocker. Children have parent-child relationship to Epic.

```bash
bd dep add test-ac0 test-n9r --type blocks  # Epic blocked by Blocker
```

**Result:** Children are NOT ready, even though they have no direct blocks dependency!

```
📋 Ready work (1 issues with no blockers):
1. [P1] test-n9r: Blocker task
```

**Key insight:** `bd ready` considers an issue blocked if:
1. It has a direct `blocks` dependency on an open issue, OR
2. Its parent (via parent-child) is blocked (transitively)

---

## Finding 4: Explicit `status: blocked` Does NOT Propagate to Children

**Scenario:** Epic has `status: blocked` set explicitly (not via dependency).

```bash
bd update test-ac0 --status blocked
```

**Result:** Children are still ready!

```
📋 Ready work (2 issues with no blockers):
1. [P2] test-3q9: Child task A
2. [P2] test-c7u: Child task B
```

**Key insight:** Blocking propagation through parent-child only works for `blocks` type dependencies, not explicit status.

---

## Finding 5: Closing a Blocker Unblocks Dependents

**Scenario:** Blocker is closed.

```bash
bd close test-n9r
```

**Result:** All issues that were blocked by it (directly or transitively) become ready.

```
📋 Ready work (3 issues with no blockers):
1. [P2] test-ac0: Epic: Parent task
2. [P2] test-3q9: Child task A
3. [P2] test-c7u: Child task B
```

---

## Summary: `bd ready` Algorithm

An issue is considered **blocked** (not ready) if ANY of these are true:

1. **Direct blocks dependency:** The issue has a `blocks` dependency on an open issue
2. **Transitive blocks via parent-child:** The issue has a `parent-child` dependency on an issue that is blocked (recursively)

An issue is **NOT** considered blocked just because:
- Its `status` field is set to "blocked" (explicit status doesn't propagate)
- It's an epic with blocked children (children don't block parents)

---

## Implications for beads-tui

### Current TUI Behavior (Incorrect)

The TUI only considers direct `blocks` dependencies when categorizing issues. It does NOT check if a parent is blocked.

### Required Fix

The TUI's `categorizeIssues()` function in `internal/state/state.go` needs to:

1. First pass: Mark all issues with direct `blocks` dependencies on open issues as blocked
2. Second pass: Propagate blocked status through `parent-child` relationships (children of blocked parents are also blocked)
3. Repeat until no changes (for deep hierarchies)

### Display Fix

The dependency display should be clearer:
- Currently shows: `• blocks tui-xyz` (ambiguous)
- Should show: `• blocked by tui-xyz` or `• depends on tui-xyz [blocks]`

The key is that from the perspective of the issue being viewed, `blocks` means "this issue is blocked BY the target", not "this issue blocks the target".

---

## Finding 6: Transitive Blocking Through Deep Hierarchies

**Scenario:** Grandchild -> parent-child -> Child A -> parent-child -> Epic -> blocks -> Blocker

```bash
bd create "Grandchild task" -p 2  # test-rt7
bd dep add test-rt7 test-3q9 --type parent-child  # Grandchild is child of Child A
```

**Result:** Grandchild is blocked even though it's 3 levels away from the blocker.

```
📋 Ready work (1 issues with no blockers):
1. [P1] test-n9r: Blocker task
```

**Key insight:** Blocking propagates to arbitrary depth through parent-child chains.

---

## Finding 7: "related" Dependencies Do NOT Block

**Command:** `bd dep add A B --type related`

**Meaning:** "A is related to B" (informational link)

```bash
bd dep add test-930 test-n9r --type related
```

**Result:** The related task remains ready. `related` is purely organizational.

---

## Finding 8: "discovered-from" Dependencies Do NOT Block

**Command:** `bd dep add A B --type discovered-from`

**Meaning:** "A was discovered from B" (provenance tracking)

```bash
bd dep add test-cn1 test-n9r --type discovered-from
```

**Result:** The discovered task remains ready. `discovered-from` is purely organizational.

---

## Complete Dependency Type Summary

| Type | Blocks? | Propagates through parent-child? | Purpose |
|------|---------|----------------------------------|---------|
| `blocks` | **YES** | YES (to children) | Task ordering / prerequisites |
| `parent-child` | NO | N/A (is the propagation mechanism) | Hierarchical organization |
| `related` | NO | NO | Informational cross-reference |
| `discovered-from` | NO | NO | Provenance / audit trail |

---

## Test Commands Reference

```bash
# Create issues
bd create "Title" --type epic -p 2

# Add blocks dependency (A is blocked by B)
bd dep add A B --type blocks

# Add parent-child (A is child of B)
bd dep add A B --type parent-child

# Check ready issues
bd ready

# Show issue with dependencies
bd show ISSUE_ID

# Update status
bd update ISSUE_ID --status blocked

# Close issue
bd close ISSUE_ID
```
