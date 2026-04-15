'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Folder as FolderIcon,
  FolderOpen,
  ChevronRight,
  ChevronDown,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  FileText,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';
import type { Folder } from '@/lib/api/types';
import {
  listAllFolders,
  createFolder,
  renameFolder,
  deleteFolder,
  moveDocument,
} from '@/lib/api/folders';
import { toast } from 'sonner';

interface FolderTreeProps {
  selectedFolderId?: string;
  onFolderSelect?: (folderId: string | null) => void;
  documents?: Array<{ id: string; title: string; folder_id?: string }>;
  onDocumentClick?: (docId: string) => void;
  onRefresh?: () => void;
}

interface FolderNode extends Folder {
  children: FolderNode[];
  isOpen?: boolean;
}

export function FolderExplorer({
  selectedFolderId,
  onFolderSelect,
  documents = [],
  onDocumentClick,
  onRefresh,
}: FolderTreeProps) {
  const [folders, setFolders] = useState<FolderNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const [showNewFolderInput, setShowNewFolderInput] = useState(false);
  const [newFolderName, setNewFolderName] = useState('');
  const [newFolderParentId, setNewFolderParentId] = useState<string | null>(null);
  const [draggedDocId, setDraggedDocId] = useState<string | null>(null);

  const loadFolders = useCallback(async () => {
    try {
      const response = await listAllFolders();
      const folderMap = new Map<string, FolderNode>();
      const roots: FolderNode[] = [];

      response.folders.forEach((f) => {
        folderMap.set(f.id, { ...f, children: [], isOpen: expandedIds.has(f.id) });
      });

      folderMap.forEach((node) => {
        if (node.parent_id && folderMap.has(node.parent_id)) {
          folderMap.get(node.parent_id)!.children.push(node);
        } else if (!node.parent_id) {
          roots.push(node);
        }
      });

      setFolders(roots);
    } catch {
      toast.error('Failed to load folders');
    } finally {
      setLoading(false);
    }
  }, [expandedIds]);

  useEffect(() => {
    loadFolders();
  }, [loadFolders]);

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
      setNewFolderName('');
      setShowNewFolderInput(false);
      setNewFolderParentId(null);
      toast.success('Folder created');
      loadFolders();
      onRefresh?.();
    } catch {
      toast.error('Failed to create folder');
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
      toast.success('Folder renamed');
      loadFolders();
      onRefresh?.();
    } catch {
      toast.error('Failed to rename folder');
    }
  };

  const handleDeleteFolder = async (folderId: string) => {
    try {
      await deleteFolder(folderId);
      toast.success('Folder deleted');
      loadFolders();
      onRefresh?.();
    } catch {
      toast.error('Failed to delete folder');
    }
  };

  const handleDragOver = (e: React.DragEvent, _folderId: string) => {
    e.preventDefault();
    e.currentTarget.setAttribute('data-drop-target', 'true');
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.currentTarget.removeAttribute('data-drop-target');
  };

  const handleDrop = async (e: React.DragEvent, folderId: string) => {
    e.preventDefault();
    e.currentTarget.removeAttribute('data-drop-target');
    const docId = e.dataTransfer.getData('text/plain');
    if (!docId) return;

    try {
      await moveDocument(docId, folderId);
      toast.success('Document moved');
      onRefresh?.();
    } catch {
      toast.error('Failed to move document');
    }
  };

  const handleDocumentDragStart = (e: React.DragEvent, docId: string) => {
    setDraggedDocId(docId);
    e.dataTransfer.setData('text/plain', docId);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDocumentDragEnd = () => {
    setDraggedDocId(null);
  };

  const renderFolder = (folder: FolderNode, depth: number = 0) => {
    const hasChildren = folder.children.length > 0;
    const isExpanded = expandedIds.has(folder.id);
    const isEditing = editingId === folder.id;
    const isSelected = selectedFolderId === folder.id;

    return (
      <div key={folder.id}>
        <div
          className={cn(
            'group flex items-center gap-1 rounded-md px-2 py-1.5 cursor-pointer transition-colors',
            isSelected && 'bg-primary/10 text-primary',
            !isSelected && 'hover:bg-accent'
          )}
          style={{ paddingLeft: `${depth * 16 + 8}px` }}
          onClick={() => {
            onFolderSelect?.(folder.id);
            if (hasChildren) toggleExpand(folder.id);
          }}
          onDragOver={(e) => handleDragOver(e, folder.id)}
          onDragLeave={handleDragLeave}
          onDrop={(e) => handleDrop(e, folder.id)}
          data-folder-id={folder.id}
        >
          {hasChildren ? (
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
                if (e.key === 'Enter') handleRenameFolder(folder.id);
                if (e.key === 'Escape') setEditingId(null);
              }}
              autoFocus
              onClick={(e) => e.stopPropagation()}
            />
          ) : (
            <span className="flex-1 truncate text-sm">{folder.name}</span>
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
                Rename
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  setNewFolderParentId(folder.id);
                  setShowNewFolderInput(true);
                }}
              >
                <Plus className="mr-2 h-4 w-4" />
                New Subfolder
              </DropdownMenuItem>
              <DropdownMenuItem
                className="text-destructive"
                onClick={(e) => {
                  e.stopPropagation();
                  handleDeleteFolder(folder.id);
                }}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {hasChildren && isExpanded && (
          <div>
            {folder.children.map((child) => renderFolder(child, depth + 1))}
          </div>
        )}

        {showNewFolderInput && newFolderParentId === folder.id && (
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
                if (e.key === 'Enter') handleCreateFolder(folder.id);
                if (e.key === 'Escape') {
                  setShowNewFolderInput(false);
                  setNewFolderName('');
                  setNewFolderParentId(null);
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

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-2 border-b">
        <h3 className="font-medium text-sm">Folders</h3>
        <Button
          variant="ghost"
          size="icon-xs"
          onClick={() => {
            setNewFolderParentId(null);
            setShowNewFolderInput(true);
          }}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {loading ? (
          <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
            Loading...
          </div>
        ) : rootFolders.length === 0 && !showNewFolderInput ? (
          <div className="flex flex-col items-center justify-center py-8 text-sm text-muted-foreground">
            <FolderIcon className="h-8 w-8 mb-2 opacity-50" />
            <p>No folders yet</p>
            <Button
              variant="link"
              size="sm"
              className="mt-1"
              onClick={() => setShowNewFolderInput(true)}
            >
              Create one
            </Button>
          </div>
        ) : (
          <div className="space-y-0.5">
            {rootFolders.map((folder) => renderFolder(folder))}

            {showNewFolderInput && newFolderParentId === null && (
              <div className="flex items-center gap-1 px-2 py-1">
                <FolderIcon className="h-4 w-4 shrink-0 text-primary" />
                <Input
                  className="h-6 flex-1"
                  placeholder="Folder name..."
                  value={newFolderName}
                  onChange={(e) => setNewFolderName(e.target.value)}
                  onBlur={() => handleCreateFolder(null)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleCreateFolder(null);
                    if (e.key === 'Escape') {
                      setShowNewFolderInput(false);
                      setNewFolderName('');
                    }
                  }}
                  autoFocus
                />
              </div>
            )}
          </div>
        )}
      </div>

      {documents.length > 0 && (
        <div className="border-t p-2">
          <h4 className="mb-2 px-2 text-xs font-medium text-muted-foreground uppercase">
            Documents
          </h4>
          <div className="space-y-0.5 max-h-48 overflow-y-auto">
            {documents.map((doc) => (
              <div
                key={doc.id}
                className={cn(
                  'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm cursor-grab transition-colors',
                  draggedDocId === doc.id && 'opacity-50',
                  'hover:bg-accent'
                )}
                draggable
                onDragStart={(e) => handleDocumentDragStart(e, doc.id)}
                onDragEnd={handleDocumentDragEnd}
                onClick={() => onDocumentClick?.(doc.id)}
              >
                <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                <span className="truncate">{doc.title}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export { type FolderNode };
