"use client";

import { useState, useEffect, useCallback } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import {
  Folder as FolderIcon,
  FolderOpen,
  FileText,
  ChevronRight,
  ChevronDown,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  Upload,
  Loader2,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import type { Folder } from "@/lib/api/types";
import type { Document } from "@/lib/api/types";
import {
  listAllFolders,
  createFolder,
  renameFolder,
  deleteFolder,
  moveDocument,
} from "@/lib/api/folders";
import { listDocuments } from "@/lib/api/documents";
import { toast } from "sonner";

interface FolderNode extends Folder {
  children: FolderNode[];
  isOpen?: boolean;
  documents: Document[];
}

export default function FileBrowserPage() {
  const t = useTranslations("documents");
  const tCommon = useTranslations("common");
  const tFolder = useTranslations("folder");
  const [folders, setFolders] = useState<FolderNode[]>([]);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedFolderId, setSelectedFolderId] = useState<string | null>(null);
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState("");
  const [showNewFolderInput, setShowNewFolderInput] = useState(false);
  const [newFolderName, setNewFolderName] = useState("");
  const [draggedDocId, setDraggedDocId] = useState<string | null>(null);
  

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [foldersRes, docsRes] = await Promise.all([
        listAllFolders(),
        listDocuments({ limit: 1000 }),
      ]);

      const folderMap = new Map<string, FolderNode>();
      const roots: FolderNode[] = [];

      foldersRes.folders.forEach((f) => {
        folderMap.set(f.id, {
          ...f,
          children: [],
          isOpen: expandedIds.has(f.id),
          documents: [],
        });
      });

      folderMap.forEach((node) => {
        if (node.parent_id && folderMap.has(node.parent_id)) {
          folderMap.get(node.parent_id)!.children.push(node);
        } else if (!node.parent_id) {
          roots.push(node);
        }
      });

      docsRes.documents.forEach((doc) => {
        if (doc.folder_id && folderMap.has(doc.folder_id)) {
          folderMap.get(doc.folder_id)!.documents.push(doc);
        }
      });

      setFolders(roots);
      setDocuments(docsRes.documents);
    } catch {
      toast.error("Failed to load data");
    } finally {
      setLoading(false);
    }
  }, [expandedIds]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const toggleExpand = (folderId: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(folderId)) {
        next.delete(folderId);
      } else {
        next.add(folderId);
      }
      return next;
    });
  };

  const handleCreateFolder = async (parentId: string | null = null) => {
    if (!newFolderName.trim()) return;
    try {
      await createFolder(newFolderName.trim(), parentId || undefined);
      setNewFolderName("");
      setShowNewFolderInput(false);
      toast.success("Folder created");
      loadData();
    } catch {
      toast.error("Failed to create folder");
    }
  };

  const handleRenameFolder = async (folderId: string) => {
    if (!editingName.trim()) {
      setEditingId(null);
      return;
    }
    try {
      await renameFolder(folderId, editingName.trim());
      setEditingId(null);
      toast.success("Folder renamed");
      loadData();
    } catch {
      toast.error("Failed to rename folder");
    }
  };

  const handleDeleteFolder = async (folderId: string) => {
    try {
      await deleteFolder(folderId);
      toast.success("Folder deleted");
      loadData();
    } catch {
      toast.error("Failed to delete folder");
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.currentTarget.setAttribute("data-drop-target", "true");
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.currentTarget.removeAttribute("data-drop-target");
  };

  const handleDrop = async (e: React.DragEvent, folderId: string) => {
    e.preventDefault();
    e.currentTarget.removeAttribute("data-drop-target");
    const docId = e.dataTransfer.getData("text/plain");
    if (!docId) return;

    try {
      await moveDocument(docId, folderId);
      toast.success("Document moved");
      loadData();
    } catch {
      toast.error("Failed to move document");
    }
  };

  const handleDocumentDragStart = (e: React.DragEvent, docId: string) => {
    setDraggedDocId(docId);
    e.dataTransfer.setData("text/plain", docId);
    e.dataTransfer.effectAllowed = "move";
  };

  const handleDocumentDragEnd = () => {
    setDraggedDocId(null);
  };

  const renderFolder = (folder: FolderNode, depth: number = 0) => {
    const hasChildren = folder.children.length > 0;
    const hasDocuments = folder.documents.length > 0;
    const isExpanded = expandedIds.has(folder.id);
    const isEditing = editingId === folder.id;
    const isSelected = selectedFolderId === folder.id;

    return (
      <div key={folder.id}>
        <div
          className={cn(
            "group flex items-center gap-1 rounded-md px-2 py-1.5 cursor-pointer transition-colors",
            isSelected && "bg-primary/10 text-primary",
            !isSelected && "hover:bg-accent",
          )}
          style={{ paddingLeft: `${depth * 16 + 8}px` }}
          onClick={() => {
            setSelectedFolderId(folder.id);
            if (hasChildren) toggleExpand(folder.id);
          }}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={(e) => handleDrop(e, folder.id)}
        >
          {hasChildren || hasDocuments ? (
            isExpanded ? (
              <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
            )
          ) : (
            <span className="w-4" />
          )}

          {isExpanded ? (
            <FolderOpen className="h-4 w-4 shrink-0 text-primary" />
          ) : (
            <FolderIcon className="h-4 w-4 shrink-0 text-primary" />
          )}

          {isEditing ? (
            <Input
              className="h-6 flex-1"
              value={editingName}
              onChange={(e) => setEditingName(e.target.value)}
              onBlur={() => handleRenameFolder(folder.id)}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleRenameFolder(folder.id);
                if (e.key === "Escape") setEditingId(null);
              }}
              autoFocus
              onClick={(e) => e.stopPropagation()}
            />
          ) : (
            <span className="flex-1 truncate text-sm">
              {folder.name}
              {hasDocuments && (
                <span className="ml-2 text-xs text-muted-foreground">
                  ({folder.documents.length})
                </span>
              )}
            </span>
          )}

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon-xs"
                className="opacity-0 group-hover:opacity-100"
                onClick={(e) => e.stopPropagation()}
              >
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  setEditingId(folder.id);
                  setEditingName(folder.name);
                }}
              >
                <Pencil className="mr-2 h-4 w-4" />
                {tCommon("rename")}
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  setShowNewFolderInput(true);
                  setNewFolderName("");
                }}
              >
                <Plus className="mr-2 h-4 w-4" />
                {tFolder("newSubfolder")}
              </DropdownMenuItem>
              <DropdownMenuItem
                className="text-destructive"
                onClick={(e) => {
                  e.stopPropagation();
                  handleDeleteFolder(folder.id);
                }}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                {tCommon("delete")}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {hasChildren && isExpanded && (
          <div>
            {folder.children.map((child) => renderFolder(child, depth + 1))}
          </div>
        )}

        {hasDocuments && isExpanded && (
          <div style={{ paddingLeft: `${(depth + 1) * 16 + 24}px` }}>
            {folder.documents.map((doc) => (
              <div
                key={doc.id}
                className={cn(
                  "group flex items-center gap-2 rounded-md px-2 py-1.5 cursor-grab transition-colors",
                  draggedDocId === doc.id && "opacity-50",
                  "hover:bg-accent",
                )}
                draggable
                onDragStart={(e) => handleDocumentDragStart(e, doc.id)}
                onDragEnd={handleDocumentDragEnd}
              >
                <Link
                  href={`/documents/${doc.id}`}
                  className="flex flex-1 items-center gap-2"
                  onClick={(e) => e.stopPropagation()}
                >
                  <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <span className="truncate text-sm">{doc.title}</span>
                </Link>
              </div>
            ))}
          </div>
        )}

        {showNewFolderInput && !isEditing && (
          <div
            className="flex items-center gap-1 px-2 py-1"
            style={{ paddingLeft: `${(depth + 1) * 16 + 8}px` }}
          >
            <FolderIcon className="h-4 w-4 shrink-0 text-primary" />
            <Input
              className="h-6 flex-1"
              placeholder="Folder name..."
              value={newFolderName}
              onChange={(e) => setNewFolderName(e.target.value)}
              onBlur={() => {
                handleCreateFolder(folder.id);
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleCreateFolder(folder.id);
                if (e.key === "Escape") {
                  setShowNewFolderInput(false);
                  setNewFolderName("");
                }
              }}
              autoFocus
            />
          </div>
        )}
      </div>
    );
  };

  const rootFolders = folders.filter((f) => !f.parent_id);

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">{t("fileBrowser")}</h1>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => loadData()}
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            {tCommon("refresh")}
          </Button>
          <Button asChild>
            <Link href="/documents/upload">
              <Upload className="mr-2 h-4 w-4" />
              {t("upload")}
            </Link>
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-1">
          <CardContent className="p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium text-sm">{tFolder("folders")}</h3>
              <Button
                variant="ghost"
                size="icon-xs"
                onClick={() => {
                  setShowNewFolderInput(true);
                  setNewFolderName("");
                }}
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>

            <div className="space-y-0.5">
              {rootFolders.map((folder) => renderFolder(folder))}

              {showNewFolderInput && (
                <div className="flex items-center gap-1 px-2 py-1">
                  <FolderIcon className="h-4 w-4 shrink-0 text-primary" />
                  <Input
                    className="h-6 flex-1"
                    placeholder="Folder name..."
                    value={newFolderName}
                    onChange={(e) => setNewFolderName(e.target.value)}
                    onBlur={() => handleCreateFolder(null)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") handleCreateFolder(null);
                      if (e.key === "Escape") {
                        setShowNewFolderInput(false);
                        setNewFolderName("");
                      }
                    }}
                    autoFocus
                  />
                </div>
              )}

              {rootFolders.length === 0 && !showNewFolderInput && (
                <div className="flex flex-col items-center justify-center py-8 text-sm text-muted-foreground">
                  <FolderIcon className="h-8 w-8 mb-2 opacity-50" />
                  <p>{tFolder("noFolders")}</p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardContent className="p-4">
            <h3 className="font-medium text-sm mb-4">
              {selectedFolderId ? `Documents in folder` : "All Documents"}
              {selectedFolderId && (
                <Button
                  variant="link"
                  size="sm"
                  className="ml-2"
                  onClick={() => setSelectedFolderId(null)}
                >
                  {tCommon("showAll")}
                </Button>
              )}
            </h3>

            <div className="space-y-2">
              {(selectedFolderId
                ? documents.filter((d) => d.folder_id === selectedFolderId)
                : documents
              ).length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8 text-sm text-muted-foreground">
                  <FileText className="h-8 w-8 mb-2 opacity-50" />
                  <p>{t("noDocuments")}</p>
                </div>
              ) : (
                (selectedFolderId
                  ? documents.filter((d) => d.folder_id === selectedFolderId)
                  : documents
                ).map((doc) => (
                  <Link
                    key={doc.id}
                    href={`/documents/${doc.id}`}
                    className={cn(
                      "group flex items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-accent",
                      draggedDocId === doc.id && "opacity-50",
                    )}
                    draggable
                    onDragStart={(e) => handleDocumentDragStart(e, doc.id)}
                    onDragEnd={handleDocumentDragEnd}
                  >
                    <FileText className="h-5 w-5 shrink-0 text-muted-foreground" />
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{doc.title}</p>
                      <p className="text-xs text-muted-foreground">
                        {doc.doc_type} &middot;{" "}
                        {new Date(doc.created_at).toLocaleDateString()}
                      </p>
                    </div>
                    <Badge variant="outline" className="shrink-0">
                      {doc.status}
                    </Badge>
                  </Link>
                ))
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function Badge({
  variant = "outline",
  children,
  className,
}: {
  variant?: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors",
        variant === "outline" && "border-border text-foreground",
        variant === "secondary" && "bg-secondary text-secondary-foreground",
        variant === "destructive" &&
          "bg-destructive text-destructive-foreground",
        className,
      )}
    >
      {children}
    </span>
  );
}
