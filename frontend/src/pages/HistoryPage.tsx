import { useState, useEffect, useCallback } from 'react';
import { historyApi } from '@/api';
import { useCluster } from '@/hooks/use-cluster';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading, Spinner } from '@/components/ui/spinner';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import Editor, { DiffEditor } from '@monaco-editor/react';
import type { HistoryRecord, HistoryDiff } from '@/types';
import {
  Search,
  History,
  RefreshCw,
  Clock,
  User,
  FileCode,
  ChevronLeft,
  ChevronRight,
  Eye,
  Copy,
  GitCompareArrows,
} from 'lucide-react';

const RESOURCE_TYPES = ['All', 'ConfigMap', 'Deployment', 'Service', 'HPA'];
const PAGE_SIZE = 20;

export function HistoryPage() {
  const { selectedCluster, selectedNamespace } = useCluster();
  const { toast } = useToast();

  const [records, setRecords] = useState<HistoryRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [resourceType, setResourceType] = useState('All');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  // Detail dialog
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedRecord, setSelectedRecord] = useState<HistoryRecord | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  // Diff dialog
  const [diffDialogOpen, setDiffDialogOpen] = useState(false);
  const [diffData, setDiffData] = useState<HistoryDiff | null>(null);
  const [diffLoading, setDiffLoading] = useState(false);
  const [diffRecord, setDiffRecord] = useState<HistoryRecord | null>(null);

  const fetchHistory = useCallback(async () => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setLoading(true);
      const data = await historyApi.list(selectedCluster, selectedNamespace, {
        resource_type: resourceType === 'All' ? undefined : resourceType,
        resource_name: searchTerm || undefined,
        page,
        page_size: PAGE_SIZE,
      });
      setRecords(data.items || []);
      setTotal(data.total || 0);
    } catch (err) {
      setRecords([]);
      setTotal(0);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch history',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, selectedNamespace, resourceType, searchTerm, page, toast]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  const fetchDetail = async (record: HistoryRecord) => {
    try {
      setDetailLoading(true);
      setSelectedRecord(record);
      setDetailDialogOpen(true);
      const detail = await historyApi.get(selectedCluster, selectedNamespace, record.id);
      setSelectedRecord(detail);
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch detail',
        variant: 'destructive',
      });
    } finally {
      setDetailLoading(false);
    }
  };

  const handleCopyYaml = async () => {
    if (!selectedRecord?.content) return;
    try {
      await navigator.clipboard.writeText(selectedRecord.content);
      toast({
        title: 'Success',
        description: 'YAML content copied to clipboard',
      });
    } catch (err) {
      toast({
        title: 'Error',
        description: 'Failed to copy to clipboard',
        variant: 'destructive',
      });
    }
  };

  const fetchDiffWithPrevious = async (record: HistoryRecord) => {
    try {
      setDiffLoading(true);
      setDiffRecord(record);
      setDiffDialogOpen(true);
      const data = await historyApi.diffWithPrevious(selectedCluster, selectedNamespace, record.id);
      setDiffData(data);
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch diff',
        variant: 'destructive',
      });
      setDiffDialogOpen(false);
    } finally {
      setDiffLoading(false);
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const getOperationBadge = (operation: string) => {
    switch (operation) {
      case 'create':
        return <Badge variant="success">Create</Badge>;
      case 'update':
        return <Badge variant="secondary">Update</Badge>;
      case 'delete':
        return <Badge variant="destructive">Delete</Badge>;
      default:
        return <Badge variant="outline">{operation}</Badge>;
    }
  };

  const getResourceTypeBadge = (type: string) => {
    const colors: Record<string, string> = {
      ConfigMap: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
      Deployment: 'bg-green-500/10 text-green-500 border-green-500/20',
      Service: 'bg-purple-500/10 text-purple-500 border-purple-500/20',
      HPA: 'bg-orange-500/10 text-orange-500 border-orange-500/20',
    };
    return (
      <Badge variant="outline" className={colors[type] || ''}>
        {type}
      </Badge>
    );
  };

  if (!selectedCluster || !selectedNamespace) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">Please select a cluster and namespace</p>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-semibold flex items-center gap-2">
          <History className="w-7 h-7 text-primary" />
          Change History
        </h1>
        <p className="text-muted-foreground mt-1">
          View resource change history in {selectedNamespace}
        </p>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 min-w-[200px] max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search by resource name..."
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value);
              setPage(1);
            }}
            className="pl-9"
          />
        </div>

        <Select
          value={resourceType}
          onValueChange={(value) => {
            setResourceType(value);
            setPage(1);
          }}
        >
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Resource Type" />
          </SelectTrigger>
          <SelectContent>
            {RESOURCE_TYPES.map((type) => (
              <SelectItem key={type} value={type}>
                {type}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button variant="outline" size="icon" onClick={fetchHistory}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* History List */}
      {loading ? (
        <Loading text="Loading history..." />
      ) : records.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <FileCode className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">No history records found</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {records.map((record) => (
            <Card key={record.id} className="hover:border-primary/30 transition-colors">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center gap-2">
                        {getResourceTypeBadge(record.resource_type)}
                        <span className="font-medium">{record.resource_name}</span>
                        {getOperationBadge(record.operation)}
                      </div>
                      <div className="flex items-center gap-4 text-sm text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <Clock className="w-3.5 h-3.5" />
                          {new Date(record.created_at).toLocaleString()}
                        </span>
                        <span className="flex items-center gap-1">
                          <User className="w-3.5 h-3.5" />
                          {record.operator || 'system'}
                        </span>
                        <span>Version: {record.version}</span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">#{record.id}</Badge>
                    {record.operation !== 'delete' && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => fetchDiffWithPrevious(record)}
                      >
                        <GitCompareArrows className="w-4 h-4 mr-1" />
                        Diff
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => fetchDetail(record)}
                    >
                      <Eye className="w-4 h-4 mr-1" />
                      View
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * PAGE_SIZE + 1} to {Math.min(page * PAGE_SIZE, total)} of {total} records
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <span className="text-sm">
              Page {page} of {totalPages}
            </span>
            <Button
              variant="outline"
              size="icon"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Detail Dialog */}
      <Dialog open={detailDialogOpen} onOpenChange={setDetailDialogOpen}>
        <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              {selectedRecord && getResourceTypeBadge(selectedRecord.resource_type)}
              {selectedRecord?.resource_name}
              {selectedRecord && getOperationBadge(selectedRecord.operation)}
            </DialogTitle>
            <DialogDescription>
              Version {selectedRecord?.version} • {selectedRecord?.created_at && new Date(selectedRecord.created_at).toLocaleString()}
            </DialogDescription>
          </DialogHeader>

          {/* Record Info */}
          <div className="flex items-center gap-4 py-2 px-3 bg-muted rounded-md text-sm">
            <div className="flex items-center gap-1.5">
              <User className="w-4 h-4 text-muted-foreground" />
              <span className="text-muted-foreground">Operator:</span>
              <span className="font-medium">{selectedRecord?.operator || 'system'}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <Clock className="w-4 h-4 text-muted-foreground" />
              <span className="text-muted-foreground">Time:</span>
              <span className="font-medium">
                {selectedRecord?.created_at && new Date(selectedRecord.created_at).toLocaleString()}
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <FileCode className="w-4 h-4 text-muted-foreground" />
              <span className="text-muted-foreground">ID:</span>
              <span className="font-medium">#{selectedRecord?.id}</span>
            </div>
          </div>

          {/* YAML Content */}
          <div className="flex-1 min-h-0 border rounded-md overflow-hidden">
            {detailLoading ? (
              <div className="flex items-center justify-center h-full">
                <Spinner size="lg" />
              </div>
            ) : (
              <Editor
                height="100%"
                defaultLanguage="yaml"
                value={selectedRecord?.content || '# No content available'}
                theme="vs-dark"
                options={{
                  readOnly: true,
                  minimap: { enabled: false },
                  fontSize: 13,
                  fontFamily: 'Fira Code, monospace',
                  lineNumbers: 'on',
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  tabSize: 2,
                  wordWrap: 'on',
                }}
              />
            )}
          </div>

          {/* Actions */}
          <div className="flex items-center justify-between pt-2">
            <p className="text-xs text-muted-foreground">
              Click copy to copy YAML content to clipboard
            </p>
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={() => setDetailDialogOpen(false)}>
                Close
              </Button>
              <Button
                onClick={handleCopyYaml}
                disabled={!selectedRecord?.content}
              >
                <Copy className="w-4 h-4 mr-1" />
                Copy YAML
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Diff Dialog */}
      <Dialog open={diffDialogOpen} onOpenChange={setDiffDialogOpen}>
        <DialogContent className="max-w-6xl h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <GitCompareArrows className="w-5 h-5" />
              {diffRecord && getResourceTypeBadge(diffRecord.resource_type)}
              {diffRecord?.resource_name}
              <span className="text-muted-foreground font-normal text-sm">
                {diffData?.source_version === 0
                  ? '(First version - no previous version)'
                  : `Version diff`}
              </span>
            </DialogTitle>
            <DialogDescription>
              {diffData?.source_version === 0
                ? `Current: Version ${diffRecord?.version}`
                : `Previous version → Current version (v${diffRecord?.version})`}
            </DialogDescription>
          </DialogHeader>

          <div className="flex-1 min-h-0 border rounded-md overflow-hidden">
            {diffLoading ? (
              <div className="flex items-center justify-center h-full">
                <Spinner size="lg" />
              </div>
            ) : (
              <DiffEditor
                height="100%"
                language="yaml"
                original={diffData?.source_content || ''}
                modified={diffData?.target_content || ''}
                theme="vs-dark"
                options={{
                  readOnly: true,
                  minimap: { enabled: false },
                  fontSize: 13,
                  fontFamily: 'Fira Code, monospace',
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  renderSideBySide: true,
                }}
              />
            )}
          </div>

          <div className="flex items-center justify-end pt-2">
            <Button variant="outline" onClick={() => setDiffDialogOpen(false)}>
              Close
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
