import { useState, useEffect } from 'react';
import Editor from '@monaco-editor/react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';

interface YamlEditorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  initialYaml: string;
  onSave: (yaml: string) => Promise<void>;
  saving?: boolean;
  readOnly?: boolean;
  saveButtonText?: string;
}

export function YamlEditorDialog({
  open,
  onOpenChange,
  title,
  description,
  initialYaml,
  onSave,
  saving = false,
  readOnly = false,
  saveButtonText,
}: YamlEditorDialogProps) {
  const [yaml, setYaml] = useState(initialYaml);

  useEffect(() => {
    setYaml(initialYaml);
  }, [initialYaml, open]);

  const handleSave = async () => {
    await onSave(yaml);
  };

  const buttonText = saveButtonText || (readOnly ? 'Edit' : 'Save');

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>

        <div className="flex-1 min-h-0 border rounded-md overflow-hidden">
          <Editor
            height="100%"
            defaultLanguage="yaml"
            value={yaml}
            onChange={(value) => !readOnly && setYaml(value || '')}
            theme="vs-dark"
            options={{
              minimap: { enabled: false },
              fontSize: 13,
              fontFamily: 'Fira Code, monospace',
              lineNumbers: 'on',
              scrollBeyondLastLine: false,
              automaticLayout: true,
              tabSize: 2,
              wordWrap: 'on',
              readOnly: readOnly,
            }}
          />
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            {readOnly ? 'Close' : 'Cancel'}
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving && <Spinner size="sm" className="mr-2" />}
            {saving ? 'Saving...' : buttonText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
