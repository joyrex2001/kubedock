# 🧭 Release Process Guide

This project uses **[Release Drafter](https://github.com/release-drafter/release-drafter)** to automate versioning and changelog generation.  
Release Drafter automatically creates or updates a **draft release** based on merged pull requests, their **labels**, and configured **categories**.

---

## 🔄 Workflow Overview

1. **Auto Labeler**
    - Pull requests are automatically labeled using the autolabeler feature of Release Drafter.
      - The configuration for the autolabeler, can be found in the `autolabeler` section of the `.github/release-drafter-config.yml` file.
      - The labels are configured based on both PR titles and changed files.
    - Labels determine both:
        - The **release note category** (e.g., *Features*, *Bug Fixes*, *Maintenance*).
          - This is configured in the `categories` section of the `.github/release-drafter-config.yml` file.
        - The **version bump** (major, minor, patch).
          - This is configured in the `version-resolver` section of the `.github/release-drafter-config.yml` file.

2. **Release Drafter**
    - A **draft release** is automatically created or updated after each merge to `master`.
    - The draft includes:
        - Suggested **next version** (based on labels).
        - **Release notes** grouped by category.

3. **Manual Review and Publish**
    - Before publishing a release:
        1. **Review PR labels**:
            - Ensure all merged PRs have correct labels c.q. are in the right category of the draft release
            - Adjust labels if needed (to fix release notes or version bump).
        2. **Delete the existing draft release** if labels were changed.
        3. **Re-run the workflow manually**:
            - Go to **Actions → Update draft release on merge → “Dispatch workflow”**.
        4. Alternatively, you can **edit the draft release manually**:
            - Change the **next tag version** if you want to override the auto-generated version.
            - Change the release notes where applicable

4. **Add Binaries**
    - Once the release is **published**, the compiled **binaries/artifacts** are uploaded as artifacts of the release. 
      This is determined via the `.github/release.yml` workflow.

---

## 💡 Tips

- If you want a PR to be excluded from the changelog, label it with one of the configured `exclude-labels`.
- If the draft release shows the wrong **next version**, you can:
    - Delete and regenerate the draft, **or**
    - Manually edit the **“Tag version”** field before publishing.
- Always verify the **release notes and version** before publishing, since they are derived automatically from labels.