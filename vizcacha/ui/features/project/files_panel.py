"""Files panel: the open folder as a tree, without .git, binaries and Delve leftovers."""

from pathlib import Path

from PyQt5.QtCore import QModelIndex, QSortFilterProxyModel, pyqtSignal
from PyQt5.QtWidgets import QFileSystemModel, QLabel, QTreeView, QVBoxLayout, QWidget

HIDDEN_NAMES = frozenset({".git"})
HIDDEN_PREFIXES = ("__debug_bin",)
BINARY_SUFFIXES = frozenset({".exe", ".dll", ".so", ".dylib", ".o", ".a", ".test", ".out"})
OPENABLE_SUFFIXES = frozenset({".go", ".mod", ".sum", ".txt", ".md"})
METADATA_COLUMNS = (1, 2, 3)  # size, type, date modified


def is_hidden_entry(name: str, is_dir: bool, is_executable: bool) -> bool:
    """True for entries a beginner should not see in the Files panel.

    ``is_executable`` matters for files without extension (Go binaries on Linux/macOS).
    """
    if name in HIDDEN_NAMES or name.startswith(HIDDEN_PREFIXES):
        return True
    if is_dir:
        return False
    suffix = Path(name).suffix.lower()
    return suffix in BINARY_SUFFIXES or (not suffix and is_executable)


def is_openable(path: Path) -> bool:
    return path.suffix.lower() in OPENABLE_SUFFIXES


class ProjectFilesFilter(QSortFilterProxyModel):
    def filterAcceptsRow(self, source_row: int, source_parent: QModelIndex) -> bool:  # noqa: N802
        model = self.sourceModel()
        info = model.fileInfo(model.index(source_row, 0, source_parent))
        executable = info.isFile() and info.isExecutable()
        return not is_hidden_entry(info.fileName(), info.isDir(), executable)


class FilesPanel(QWidget):
    file_activated = pyqtSignal(object)  # Path of a double-clicked text file

    def __init__(self) -> None:
        super().__init__()
        self.model = QFileSystemModel(self)
        self.proxy = ProjectFilesFilter(self)
        self.proxy.setSourceModel(self.model)
        self.tree = QTreeView()
        self.tree.setObjectName("project_files_tree")
        self.tree.setModel(self.proxy)
        self.tree.setHeaderHidden(True)
        for column in METADATA_COLUMNS:
            self.tree.hideColumn(column)
        self.tree.doubleClicked.connect(self._on_double_clicked)
        self.folder_label = QLabel()
        layout = QVBoxLayout(self)
        layout.setContentsMargins(0, 0, 0, 0)
        layout.addWidget(self.folder_label)
        layout.addWidget(self.tree)
        self.folder: Path | None = None

    def set_folder(self, folder: Path) -> None:
        self.folder = Path(folder)
        self.folder_label.setText(self.folder.name or str(self.folder))
        self.folder_label.setToolTip(str(self.folder))
        source_root = self.model.setRootPath(str(self.folder))
        self.tree.setRootIndex(self.proxy.mapFromSource(source_root))

    def path_at(self, proxy_index: QModelIndex) -> Path:
        return Path(self.model.filePath(self.proxy.mapToSource(proxy_index)))

    def _on_double_clicked(self, proxy_index: QModelIndex) -> None:
        path = self.path_at(proxy_index)
        if path.is_file() and is_openable(path):
            self.file_activated.emit(path)
