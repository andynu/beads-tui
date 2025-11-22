# SBATCH Directive Frequency Analysis Guide

A comprehensive guide for Slurm cluster administrators to analyze the most commonly used SBATCH directives in their environment.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Method 1: Query Historical Job Scripts](#method-1-query-historical-job-scripts)
- [Method 2: Parse User Submit Directories](#method-2-parse-user-submit-directories)
- [Method 3: Query Slurm's Job Script Cache](#method-3-query-slurms-job-script-cache)
- [Method 4: Query Active/Recent Jobs via sacct](#method-4-query-activerecent-jobs-via-sacct)
- [Method 5: Comprehensive Analysis Script](#method-5-comprehensive-analysis-script)
- [Best Practices](#best-practices)
- [Interpreting Results](#interpreting-results)
- [Privacy and Security Considerations](#privacy-and-security-considerations)

---

## Overview

This guide provides multiple approaches to extract and analyze SBATCH directive usage patterns across your Slurm cluster. The goal is to understand which directives users employ most frequently, helping you optimize defaults, documentation, and cluster configuration.

## Prerequisites

- Administrative access to the Slurm cluster
- Access to user home directories (for comprehensive analysis)
- Slurm accounting enabled (optional, for historical queries)
- Basic familiarity with bash scripting

---

## Method 1: Query Historical Job Scripts

If your Slurm installation has accounting enabled, you can query the database for historical job submissions.

### Command

```bash
sacct -a --format=JobID,Submit,BatchScript --state=COMPLETED,RUNNING,PENDING -S now-30days | \
  grep -v "^---" | \
  awk '{print $3}' | \
  grep "^#SBATCH" | \
  sort | uniq -c | sort -rn
```

### Notes

- Requires Slurm accounting to be enabled
- `BatchScript` field may not be available in all Slurm versions
- Adjust time window with `-S` flag (e.g., `now-90days`, `2024-01-01`)
- Only captures jobs that have been submitted, not all user scripts

### Example Output

```
    156 --partition=gpu
    142 --time=24:00:00
    128 --mem=32G
     98 --cpus-per-task=4
     87 --nodes=1
```

---

## Method 2: Parse User Submit Directories

This is the most comprehensive approach, scanning user directories for job script files.

### Basic Command

```bash
find /home -type f \( -name "*.sh" -o -name "*.slurm" -o -name "*.sbatch" \) 2>/dev/null | \
  xargs grep -h "^#SBATCH" 2>/dev/null | \
  sed 's/^#SBATCH[[:space:]]*//' | \
  cut -d' ' -f1 | \
  sort | uniq -c | sort -rn
```

### Extended Version (Multiple Directories)

```bash
find /home /project /scratch -type f \( -name "*.sh" -o -name "*.slurm" -o -name "*.sbatch" \) 2>/dev/null | \
  xargs grep -h "^#SBATCH" 2>/dev/null | \
  sed 's/^#SBATCH[[:space:]]*//' | \
  cut -d' ' -f1 | \
  sort | uniq -c | sort -rn | head -30
```

### Breakdown

1. `find` - Searches for script files with common extensions
2. `xargs grep` - Extracts lines starting with `#SBATCH`
3. `sed` - Removes the `#SBATCH` prefix
4. `cut` - Extracts just the directive name (before `=` or space)
5. `sort | uniq -c` - Counts frequency
6. `sort -rn` - Sorts by frequency (descending)

### Customization

Adjust the search paths based on your environment:

```bash
SEARCH_PATHS="/home /project /scratch /work /data"
find $SEARCH_PATHS -type f \( -name "*.sh" -o -name "*.slurm" -o -name "*.sbatch" \) ...
```

---

## Method 3: Query Slurm's Job Script Cache

Slurm stores submitted job scripts in its spool directory. This provides recently submitted jobs.

### Locate Spool Directory

```bash
SPOOL_DIR=$(scontrol show config | grep SlurmdSpoolDir | awk '{print $3}')
echo "Spool directory: $SPOOL_DIR"
```

### Query Job Scripts

```bash
find $SPOOL_DIR -name "job*" -type f 2>/dev/null | \
  xargs grep -h "^#SBATCH" 2>/dev/null | \
  sed 's/^#SBATCH[[:space:]]*//' | \
  cut -d' ' -f1 | \
  sort | uniq -c | sort -rn
```

### Notes

- Spool directory location varies by installation
- Common paths: `/var/spool/slurmd`, `/var/spool/slurm`, `/tmp/slurmd`
- Only contains recent/active jobs (cache is periodically cleaned)
- Requires root/admin access to spool directory

---

## Method 4: Query Active/Recent Jobs via sacct

Extract common resource request patterns from the accounting database.

### Command

```bash
sacct -a --format=ReqMem,ReqCPUS,Partition,TimeLimit,QOS -S now-7days --noheader | \
  sort | uniq -c | sort -rn | head -20
```

### Adjust Time Window

```bash
# Last 30 days
sacct -a --format=ReqMem,ReqCPUS,Partition,TimeLimit,QOS -S now-30days --noheader | \
  sort | uniq -c | sort -rn | head -30

# Specific date range
sacct -a --format=ReqMem,ReqCPUS,Partition,TimeLimit,QOS -S 2024-01-01 -E 2024-03-31 --noheader | \
  sort | uniq -c | sort -rn | head -30
```

### Custom Format Fields

```bash
# Focus on specific attributes
sacct -a --format=Partition,QOS,Account -S now-90days --noheader | \
  sort | uniq -c | sort -rn
```

### Available Fields

Common useful fields for analysis:
- `ReqMem` - Requested memory
- `ReqCPUS` - Requested CPUs
- `Partition` - Partition name
- `TimeLimit` - Time limit requested
- `QOS` - Quality of Service
- `Account` - Account name
- `ReqNodes` - Number of nodes requested

---

## Method 5: Comprehensive Analysis Script

This script provides a detailed breakdown of directive usage patterns.

### Script: `sbatch_stats.sh`

```bash
#!/bin/bash
# sbatch_stats.sh
# Comprehensive SBATCH directive frequency analysis

echo "=== SBATCH Directive Frequency Analysis ==="
echo ""

# Adjust these paths based on your environment
SEARCH_PATHS="/home /project /scratch"

# Optional: Limit to recently modified files (performance optimization)
# FIND_OPTS="-mtime -90"  # Files modified in last 90 days

echo "Scanning for job scripts in: $SEARCH_PATHS"
TEMP_FILE=$(mktemp)

# Find all job scripts and extract SBATCH directives
find $SEARCH_PATHS -type f \( -name "*.sh" -o -name "*.slurm" -o -name "*.sbatch" \) 2>/dev/null | \
  xargs grep -h "^#SBATCH" 2>/dev/null > "$TEMP_FILE"

TOTAL_DIRECTIVES=$(wc -l < "$TEMP_FILE")
TOTAL_SCRIPTS=$(find $SEARCH_PATHS -type f \( -name "*.sh" -o -name "*.slurm" -o -name "*.sbatch" \) 2>/dev/null | wc -l)

echo "Total job scripts found: $TOTAL_SCRIPTS"
echo "Total SBATCH directives found: $TOTAL_DIRECTIVES"
echo ""

# Overall directive frequency
echo "=== Top 20 Most Common Directives ==="
cat "$TEMP_FILE" | \
  sed 's/^#SBATCH[[:space:]]*//' | \
  cut -d'=' -f1 | \
  cut -d' ' -f1 | \
  sort | uniq -c | sort -rn | head -20

echo ""
echo "=== Partition Usage ==="
grep -o "\-\-partition=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--partition=//' | \
  sort | uniq -c | sort -rn
grep -o "\-p[[:space:]]*[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/-p[[:space:]]*//' | \
  sort | uniq -c | sort -rn

echo ""
echo "=== Time Limit Patterns ==="
grep -o "\-\-time=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--time=//' | \
  sort | uniq -c | sort -rn | head -10
grep -o "\-t[[:space:]]*[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/-t[[:space:]]*//' | \
  sort | uniq -c | sort -rn | head -10

echo ""
echo "=== Memory Request Patterns ==="
grep -o "\-\-mem=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--mem=//' | \
  sort | uniq -c | sort -rn | head -10
grep -o "\-\-mem-per-cpu=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--mem-per-cpu=//' | \
  sort | uniq -c | sort -rn | head -10

echo ""
echo "=== CPU/Node Patterns ==="
echo "CPUs per task:"
grep -o "\-\-cpus-per-task=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--cpus-per-task=//' | \
  sort | uniq -c | sort -rn | head -10
grep -o "\-c[[:space:]]*[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/-c[[:space:]]*//' | \
  sort | uniq -c | sort -rn | head -10

echo ""
echo "Nodes:"
grep -o "\-\-nodes=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--nodes=//' | \
  sort | uniq -c | sort -rn | head -10
grep -o "\-N[[:space:]]*[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/-N[[:space:]]*//' | \
  sort | uniq -c | sort -rn | head -10

echo ""
echo "Tasks per node:"
grep -o "\-\-ntasks-per-node=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--ntasks-per-node=//' | \
  sort | uniq -c | sort -rn | head -10

echo ""
echo "=== GPU Requests ==="
grep -o "\-\-gres=gpu[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--gres=//' | \
  sort | uniq -c | sort -rn

echo ""
echo "=== Job Names ==="
echo "Most common job name patterns (top 10):"
grep -o "\-\-job-name=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--job-name=//' | \
  sort | uniq -c | sort -rn | head -10
grep -o "\-J[[:space:]]*[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/-J[[:space:]]*//' | \
  sort | uniq -c | sort -rn | head -10

echo ""
echo "=== Output/Error File Patterns ==="
echo "Output files:"
grep -o "\-\-output=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--output=//' | \
  sort | uniq -c | sort -rn | head -5
echo ""
echo "Error files:"
grep -o "\-\-error=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--error=//' | \
  sort | uniq -c | sort -rn | head -5

echo ""
echo "=== Email Notification Usage ==="
grep -o "\-\-mail-type=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--mail-type=//' | \
  sort | uniq -c | sort -rn

echo ""
echo "=== Account Usage ==="
grep -o "\-\-account=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--account=//' | \
  sort | uniq -c | sort -rn
grep -o "\-A[[:space:]]*[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/-A[[:space:]]*//' | \
  sort | uniq -c | sort -rn

echo ""
echo "=== QOS (Quality of Service) Usage ==="
grep -o "\-\-qos=[^[:space:]]*" "$TEMP_FILE" | \
  sed 's/--qos=//' | \
  sort | uniq -c | sort -rn

# Cleanup
rm "$TEMP_FILE"

echo ""
echo "=== Analysis Complete ==="
```

### Usage

```bash
# Make executable
chmod +x sbatch_stats.sh

# Run the analysis
./sbatch_stats.sh

# Save output to file
./sbatch_stats.sh > sbatch_analysis_$(date +%Y%m%d).txt

# Run with custom search paths
SEARCH_PATHS="/home /work" ./sbatch_stats.sh
```

---

## Best Practices

### 1. Performance Optimization

For large filesystems, consider:

```bash
# Limit to recently modified files
find /home -type f -mtime -90 \( -name "*.sh" -o -name "*.slurm" \) ...

# Exclude certain directories
find /home -type f \( -name "*.sh" -o -name "*.slurm" \) \
  -not -path "*/archive/*" \
  -not -path "*/old/*" ...

# Use parallel processing
find /home -type f -name "*.sbatch" -print0 | \
  xargs -0 -P 8 grep -h "^#SBATCH"
```

### 2. Combine Multiple Methods

Get the most complete picture by running multiple approaches:

```bash
# Method 1: User directories (what users have written)
./sbatch_stats.sh > analysis_scripts.txt

# Method 2: sacct (what was actually submitted)
sacct -a --format=ReqMem,ReqCPUS,Partition -S now-30days --noheader | \
  sort | uniq -c | sort -rn > analysis_submitted.txt

# Compare the two for discrepancies
```

### 3. Regular Monitoring

Set up a cron job for periodic analysis:

```bash
# Add to crontab
# Run weekly analysis every Monday at 2 AM
0 2 * * 1 /path/to/sbatch_stats.sh > /var/log/slurm/sbatch_analysis_$(date +\%Y\%m\%d).txt
```

### 4. Handle Both Short and Long Forms

Remember that SBATCH directives have multiple forms:

| Short Form | Long Form |
|------------|-----------|
| `-p` | `--partition` |
| `-t` | `--time` |
| `-N` | `--nodes` |
| `-n` | `--ntasks` |
| `-c` | `--cpus-per-task` |
| `-J` | `--job-name` |
| `-A` | `--account` |

The comprehensive script handles both forms.

---

## Interpreting Results

### Common Patterns to Look For

1. **Default values being explicitly set**
   - If many users specify `--nodes=1`, consider if this should be the default

2. **Unusual resource requests**
   - Extremely large memory requests might indicate need for high-memory partition
   - Very short time limits might suggest a quick/test queue would be useful

3. **Partition distribution**
   - Heavy usage of one partition might indicate need for expansion
   - Underutilized partitions might need better documentation or reallocation

4. **Missing directives**
   - Few email notifications? Users might not know about the feature
   - No account specifications? Might need better documentation

5. **Inconsistent patterns**
   - Wide variation in similar job types suggests need for templates or examples

### Example Analysis

```
Top Directives:
  1250 --partition        # Everyone specifies partition (no good default?)
   980 --time             # Time limits widely used (good)
   875 --mem              # Memory requests common
   456 --cpus-per-task    # Moderate CPU specification
   234 --mail-type        # Low email usage (awareness issue?)
    45 --qos              # QOS rarely used (awareness or availability issue?)
```

**Insights:**
- Consider setting a default partition
- Email notifications might need better documentation
- QOS feature might be underutilized

---

## Privacy and Security Considerations

### 1. User Privacy

- Scanning user home directories may expose sensitive information
- Job scripts might contain credentials, API keys, or proprietary code
- Ensure compliance with your organization's privacy policies

### Recommendations

```bash
# Option 1: Announce the audit
echo "Informing users via email before running analysis..."

# Option 2: Limit to metadata only
# Use sacct instead of file scanning

# Option 3: Anonymize results
# Strip sensitive values from output
```

### 2. Data Retention

```bash
# Don't store full scripts, just statistics
# Avoid logging:
# - Email addresses
# - Absolute file paths
# - Job names that might contain sensitive info
```

### 3. Access Control

```bash
# Restrict access to analysis results
chmod 600 sbatch_analysis.txt
chown admin:admin sbatch_analysis.txt

# Store in secure location
mv sbatch_analysis.txt /root/slurm_reports/
```

---

## Troubleshooting

### Issue: "Permission denied" errors

```bash
# Run with appropriate privileges
sudo ./sbatch_stats.sh

# Or adjust search paths to accessible locations only
SEARCH_PATHS="/scratch /project" ./sbatch_stats.sh
```

### Issue: Script runs too slowly

```bash
# Limit search scope
find /home -maxdepth 3 -type f ...

# Use mtime to limit to recent files
find /home -type f -mtime -60 ...

# Use GNU parallel for faster processing
find /home -type f -name "*.sbatch" | parallel -j 16 grep "^#SBATCH" {} > directives.txt
```

### Issue: No results from sacct

```bash
# Check if accounting is enabled
sacctmgr show configuration

# Verify accounting database
scontrol show config | grep AccountingStorage

# Check time range
sacct -a --format=JobID,Submit -S now-7days | head
```

---

## Advanced Analysis

### Correlate with User Groups

```bash
# Group directive usage by user
for user in $(ls /home); do
  echo "User: $user"
  find /home/$user -name "*.sbatch" -exec grep "^#SBATCH" {} \; | \
    cut -d' ' -f1 | sort | uniq -c
done
```

### Temporal Analysis

```bash
# Compare directive usage over time
for month in {1..12}; do
  echo "Month: 2024-$month"
  sacct -a --format=Partition,ReqMem -S 2024-$month-01 -E 2024-$month-31 --noheader | \
    sort | uniq -c | sort -rn | head -5
done
```

### Export to CSV for Further Analysis

```bash
# Export directive frequency to CSV
echo "directive,count" > sbatch_directives.csv
find /home -name "*.sbatch" -exec grep "^#SBATCH" {} \; | \
  sed 's/^#SBATCH[[:space:]]*//' | \
  cut -d'=' -f1 | \
  sort | uniq -c | \
  awk '{print $2","$1}' >> sbatch_directives.csv
```

---

## Conclusion

This guide provides multiple approaches to analyze SBATCH directive usage on your Slurm cluster. Choose the method(s) that best fit your environment, security requirements, and analysis goals.

**Quick Start Recommendation:**
1. Start with Method 4 (sacct) - least invasive, shows actual usage
2. Run Method 5 (comprehensive script) - detailed analysis
3. Compare results and investigate discrepancies

Use the insights to:
- Optimize cluster defaults
- Improve documentation
- Identify training needs
- Plan capacity upgrades
- Create job script templates

---

## Additional Resources

- [Slurm SBATCH Documentation](https://slurm.schedmd.com/sbatch.html)
- [Slurm Accounting](https://slurm.schedmd.com/accounting.html)
- [Slurm Configuration Guide](https://slurm.schedmd.com/slurm.conf.html)

---

**Version:** 1.0
**Last Updated:** 2025-11-21
**Maintainer:** Cluster Admin Team
