# Changelog

All notable changes to **Eclat** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.2] - 2025-12-31

### 🚀 Added
- **Better Asset Interaction**: Double-clicking an asset now opens it directly in your system's default application (e.g., Blender, Photoshop, Image Viewer) instead of just showing it in the file explorer.

### 🐛 Fixed
- **Workflow Stability**: Improved CI/CD pipeline with better dependency management for Linux and Windows installers.

---

## [0.1.1] - 2025-12-31

### 🚀 Added
- **Smart Auto-Updater**: Windows users can now enjoy seamless updates. The app will automatically download and install the latest versions from GitHub.
- **Update History**: Added a new section in Settings to view release notes for all new versions, so you never miss a feature.
- **Automated Builds**: Implemented a robust CI/CD pipeline to ensure reliable releases across **Windows**, **macOS**, and **Linux**.

### 🐛 Fixed
- **Version Precision**: Fixed a bug where the update checker would sometimes show notifications for the currently installed version.
- **Installer Reliability**: Improved the Windows installer generation process to prevent build failures.
- **Cross-Platform Compatibility**: Replaced CGO-dependent image libraries with pure Go alternatives, ensuring the app runs smoothly regardless of your system configuration.

---

## [0.1.0] - 2025-12-05

### 🚀 Added
- **Asset Scanning**: Recursive scanning of directories for 3D models, textures, and images.
- **Smart Recognition**: Automatic duplicate detection using file hashing and grouping of related files.
- **Metadata Extraction**: Automatic extraction of dimensions, bit depth, and dominant color calculation.
- **Thumbnail Engine**: High-performance thumbnail generation for various formats including `.blend`, `.fbx`, `.psd`, and `.exr`.
- **Real-time Watching**: Live file system monitoring to keep your library in sync instantly.
- **Modern UI**: Fully reactive interface built with React, HeroUI, and Tailwind CSS.

---
[0.1.2]: https://github.com/Pepegakac123/eclat/releases/tag/v0.1.2
[0.1.1]: https://github.com/Pepegakac123/eclat/releases/tag/v0.1.1
[0.1.0]: https://github.com/Pepegakac123/eclat/releases/tag/v0.1.0