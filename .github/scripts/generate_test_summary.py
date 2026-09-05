#!/usr/bin/env python3
"""
generate_test_summary.py
Generates a rich GitHub Actions Job Summary ($GITHUB_STEP_SUMMARY) from JUnit XML
and Go coverage profile reports.
"""

import os
import sys
import xml.etree.ElementTree as ET
from pathlib import Path


def parse_junit(junit_file: Path):
    if not junit_file.is_file():
        return None

    try:
        tree = ET.parse(junit_file)
        root = tree.getroot()
    except Exception as e:
        return {"error": f"Failed to parse JUnit XML: {e}"}

    testsuites = []
    if root.tag == "testsuite":
        testsuites = [root]
    elif root.tag == "testsuites":
        testsuites = list(root.findall("testsuite"))
    else:
        testsuites = list(root.iter("testsuite"))

    total_tests = 0
    total_failures = 0
    total_errors = 0
    total_skipped = 0
    total_time = 0.0
    failed_cases = []

    for suite in testsuites:
        try:
            total_time += float(suite.attrib.get("time", 0.0))
        except ValueError:
            pass

        for case in suite.findall("testcase"):
            total_tests += 1
            case_name = case.attrib.get("name", "Unknown")
            classname = case.attrib.get("classname", suite.attrib.get("name", ""))
            try:
                case_time = float(case.attrib.get("time", 0.0))
            except ValueError:
                case_time = 0.0

            failure = case.find("failure")
            error = case.find("error")
            skipped = case.find("skipped")

            if failure is not None or error is not None:
                elem = failure if failure is not None else error
                if failure is not None:
                    total_failures += 1
                else:
                    total_errors += 1
                msg = elem.attrib.get("message", "")
                text = (elem.text or "").strip()
                failed_cases.append({
                    "name": case_name,
                    "classname": classname,
                    "time": case_time,
                    "message": msg,
                    "details": text,
                })
            elif skipped is not None:
                total_skipped += 1

    passed_tests = total_tests - (total_failures + total_errors + total_skipped)

    return {
        "total": total_tests,
        "passed": max(0, passed_tests),
        "failures": total_failures + total_errors,
        "skipped": total_skipped,
        "time": total_time,
        "failed_cases": failed_cases,
    }


def parse_coverage(coverage_file: Path):
    if not coverage_file.is_file():
        return None

    file_stats = {}
    total_stmts = 0
    covered_stmts = 0

    try:
        with open(coverage_file, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("mode:"):
                    continue
                # Expected format: file:start.col,end.col num_stmts count
                parts = line.split()
                if len(parts) != 3:
                    continue
                file_range, stmts_str, count_str = parts
                file_path = file_range.split(":")[0]
                try:
                    stmts = int(stmts_str)
                    count = int(count_str)
                except ValueError:
                    continue

                if file_path not in file_stats:
                    file_stats[file_path] = {"total": 0, "covered": 0}

                file_stats[file_path]["total"] += stmts
                total_stmts += stmts
                if count > 0:
                    file_stats[file_path]["covered"] += stmts
                    covered_stmts += stmts
    except Exception:
        return None

    pct = (covered_stmts / total_stmts * 100) if total_stmts > 0 else 0.0
    return {
        "total_statements": total_stmts,
        "covered_statements": covered_stmts,
        "percentage": pct,
        "files": file_stats,
    }


def generate_markdown(junit_data, coverage_data):
    md = []

    if junit_data is None:
        md.append("## ⚠️ Go Test Results Not Found\n")
        md.append("No test results were found. Please inspect prior workflow step logs for build, compilation, or runner errors.\n")
        return "\n".join(md)

    if "error" in junit_data:
        md.append(f"## ⚠️ Error Parsing Test Results\n\n{junit_data['error']}\n")
        return "\n".join(md)

    total = junit_data["total"]
    passed = junit_data["passed"]
    failures = junit_data["failures"]
    skipped = junit_data["skipped"]
    duration = junit_data["time"]

    status_badge = "✅ Passed" if failures == 0 else "❌ Failed"
    cov_str = f"{coverage_data['percentage']:.1f}%" if coverage_data and coverage_data["total_statements"] > 0 else "N/A"

    md.append("## 🧪 Go Test Summary\n")
    md.append("| Status | Total Tests | Passed | Failed | Skipped | Duration | Coverage |")
    md.append("| :---: | :---: | :---: | :---: | :---: | :---: | :---: |")
    md.append(f"| **{status_badge}** | **{total}** | {passed} | {failures} | {skipped} | {duration:.2f}s | {cov_str} |\n")

    if failures > 0 and junit_data["failed_cases"]:
        md.append(f"### ❌ Failed Tests ({len(junit_data['failed_cases'])})\n")
        for fail in junit_data["failed_cases"]:
            title = f"{fail['classname']}.{fail['name']}" if fail["classname"] else fail["name"]
            md.append("<details open>")
            md.append(f"<summary><b><code>{title}</code></b> ({fail['time']:.2f}s)</summary>\n")
            details = fail["details"] or fail["message"] or "Test failed without explicit output."
            if len(details) > 3000:
                details = details[:3000] + "\n... [truncated]"
            md.append(f"```text\n{details}\n```")
            md.append("</details>\n")

    if coverage_data and coverage_data["total_statements"] > 0:
        cov_pct = coverage_data["percentage"]
        cov_covered = coverage_data["covered_statements"]
        cov_total = coverage_data["total_statements"]

        md.append("### 📊 Code Coverage\n")
        md.append(f"**Overall Statement Coverage:** `{cov_pct:.1f}%` ({cov_covered}/{cov_total} statements)\n")

        if coverage_data["files"]:
            md.append("<details>")
            md.append("<summary><b>Coverage by File</b></summary>\n")
            md.append("| File | Statements | Covered | Coverage |")
            md.append("| :--- | :---: | :---: | :---: |")
            for file_path, stats in sorted(coverage_data["files"].items()):
                f_total = stats["total"]
                f_cov = stats["covered"]
                f_pct = (f_cov / f_total * 100) if f_total > 0 else 0.0
                short_path = file_path
                for prefix in ["cbr-worker/", "cbr-worker\\"]:
                    if short_path.startswith(prefix):
                        short_path = short_path[len(prefix):]
                md.append(f"| `{short_path}` | {f_total} | {f_cov} | {f_pct:.1f}% |")
            md.append("\n</details>\n")

    return "\n".join(md)


def main():
    junit_path = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("test-results/junit.xml")
    cov_path = Path(sys.argv[2]) if len(sys.argv) > 2 else Path("test-results/coverage.out")

    summary_target = os.environ.get("GITHUB_STEP_SUMMARY")
    output_path = Path(sys.argv[3]) if len(sys.argv) > 3 else (Path(summary_target) if summary_target else None)

    junit_data = parse_junit(junit_path)
    cov_data = parse_coverage(cov_path)
    markdown_content = generate_markdown(junit_data, cov_data)

    if output_path:
        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, "a" if output_path.exists() else "w", encoding="utf-8") as f:
            f.write(markdown_content + "\n")
    else:
        print(markdown_content)


if __name__ == "__main__":
    main()
