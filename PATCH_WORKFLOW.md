# Patch-Based Collaboration Workflow

This document describes the patch-based workflow for collaborating between local development and AI assistance (Claude/Gemini).

## Overview

Instead of sharing entire codebases, we use git patches to share only the specific changes being worked on. This approach:

- ✅ Keeps context focused and manageable
- ✅ Preserves git history and metadata
- ✅ Works with any git repository
- ✅ Enables precise change tracking
- ✅ Reduces token usage for AI interactions

## Workflow Steps

### 1. Make Local Changes

Work on your code normally:
```bash
# Edit files, add features, fix bugs
vim internal/cli/create.go
vim internal/dsl/dsl.go

# Add new files if needed
git add new_file.go
```

### 2. Generate Patch File

Create a patch file containing all your changes:
```bash
./generate_patch.sh
```

This creates `my_changes.patch` containing:
- All staged changes (`git add`ed files)
- All unstaged changes (modified but not staged)
- Combined into a single patch file

### 3. Upload Patch to AI

Upload `my_changes.patch` to Claude or Gemini with your request:
```
"Please review my_changes.patch and implement the following feature..."
```

### 4. Receive AI Response

The AI will provide `gemini_response.patch` (or similar) containing their changes.

### 5. Apply AI Changes

Apply the AI's patch to your local repository:
```bash
./apply_patch.sh gemini_response.patch
```

### 6. Review and Commit

Review the applied changes and commit if satisfied:
```bash
# Review changes
git diff
git status

# Test changes
make test
make lint

# Commit if satisfied
git add .
git commit -m "Apply AI suggestions: [description]"
```

## Script Reference

### `generate_patch.sh`

**Usage:** `./generate_patch.sh`

**What it does:**
- Checks for git repository
- Combines staged and unstaged changes
- Creates `my_changes.patch`
- Shows statistics and file list
- Provides guidance for next steps

**Output:** `my_changes.patch` ready for upload

### `apply_patch.sh`

**Usage:** `./apply_patch.sh <patch_file>`

**Example:** `./apply_patch.sh gemini_response.patch`

**What it does:**
- Validates patch file exists and has content
- Shows what files will be modified
- Warns about uncommitted changes
- Applies patch with conflict detection
- Provides post-application guidance

**Safety features:**
- Dry-run validation before applying
- 3-way merge for conflict resolution
- Clear error messages and troubleshooting

## Best Practices

### Before Generating Patches

1. **Stage important files:**
   ```bash
   git add important_file.go  # Include in patch
   ```

2. **Check what will be included:**
   ```bash
   git status  # See staged/unstaged files
   git diff HEAD  # Preview patch content
   ```

### When Uploading Patches

1. **Provide clear context:**
   ```
   "Please review my_changes.patch. I'm trying to add feature X but having trouble with Y."
   ```

2. **Specify what you want:**
   ```
   "Please fix the bug in my_changes.patch and add error handling."
   ```

### After Applying Patches

1. **Always review changes:**
   ```bash
   git diff  # See what was applied
   ```

2. **Test thoroughly:**
   ```bash
   make test
   make build-greenteagc
   ./dsl-poc create --cbu="TEST"
   ```

3. **Commit with descriptive messages:**
   ```bash
   git commit -m "Add KYC validation feature with error handling"
   ```

## Troubleshooting

### "No changes detected"

**Problem:** `generate_patch.sh` says no changes detected

**Solutions:**
```bash
git status  # Check file status
git add .   # Stage all changes
./generate_patch.sh  # Try again
```

### "Patch could not be applied"

**Problem:** `apply_patch.sh` fails to apply patch

**Causes & Solutions:**

1. **Different base commit:**
   ```bash
   git log --oneline -5  # Check recent commits
   git pull  # Update to latest
   ```

2. **Conflicting local changes:**
   ```bash
   git stash  # Stash local changes
   ./apply_patch.sh gemini_response.patch
   git stash pop  # Restore local changes
   ```

3. **File moved/deleted:**
   - Manually review the patch file
   - Apply changes manually
   - Create new patch with resolved conflicts

### "Patch validation failed"

**Problem:** Patch format issues

**Solutions:**
1. Check the patch file format
2. Ensure it was generated with `git diff`
3. Try applying with `--3way` option (script does this automatically)

## Advanced Usage

### Working with Multiple Patches

```bash
# Generate patch for specific files only
git diff HEAD -- file1.go file2.go > specific_changes.patch

# Apply multiple patches in sequence
./apply_patch.sh patch1.patch
./apply_patch.sh patch2.patch
```

### Patch File Inspection

```bash
# View patch contents
cat my_changes.patch

# See files affected
grep '^diff --git' my_changes.patch

# Count changes
git apply --stat my_changes.patch
```

### Integration with Git

```bash
# Create patch from specific commits
git format-patch HEAD~2  # Last 2 commits

# Apply patches as commits
git am < gemini_response.patch
```

## File Management

The workflow uses these files:

- `my_changes.patch` - Your local changes (upload this)
- `gemini_response.patch` - AI response (download this)
- `generate_patch.sh` - Script to create patches
- `apply_patch.sh` - Script to apply patches

**Git ignore patterns:**
```gitignore
*.patch
!*_template.patch
```

Add to `.gitignore` to avoid committing patch files (they're temporary).