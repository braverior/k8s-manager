import { useState, useEffect, useCallback } from 'react';
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
import { AlertCircle } from 'lucide-react';
import { validateYamlSyntax, validateK8sYaml } from '@/lib/yaml-validator';

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
  /** 校验模式：'syntax' 仅校验 YAML 语法，'k8s' 额外校验 K8s 必填字段，默认 'k8s' */
  validateMode?: 'syntax' | 'k8s';
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
  validateMode = 'k8s',
}: YamlEditorDialogProps) {
  const [yaml, setYaml] = useState(initialYaml);
  const [validationError, setValidationError] = useState<string | undefined>();

  useEffect(() => {
    setYaml(initialYaml);
    setValidationError(undefined);
  }, [initialYaml, open]);

  const validate = useCallback(
    (content: string) => {
      if (readOnly) return true;
      const result =
        validateMode === 'k8s'
          ? validateK8sYaml(content)
          : validateYamlSyntax(content);
      setValidationError(result.valid ? undefined : result.error);
      return result.valid;
    },
    [readOnly, validateMode]
  );

  const handleChange = (value: string | undefined) => {
    if (readOnly) return;
    const content = value || '';
    setYaml(content);
    // 有内容时才实时校验，避免空白时立刻报错
    if (content.trim()) {
      validate(content);
    } else {
      setValidationError(undefined);
    }
  };

  const handleSave = async () => {
    if (!validate(yaml)) return;
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

        <div className="flex-1 min-h-0 flex flex-col gap-2">
          <div className="flex-1 min-h-0 border rounded-md overflow-hidden">
            <Editor
              height="100%"
              defaultLanguage="yaml"
              value={yaml}
              onChange={handleChange}
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

          {validationError && (
            <div className="flex items-start gap-2 px-3 py-2 rounded-md bg-destructive/10 border border-destructive/30 text-destructive text-sm">
              <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" />
              <span>{validationError}</span>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            {readOnly ? 'Close' : 'Cancel'}
          </Button>
          <Button onClick={handleSave} disabled={saving || (!readOnly && !!validationError)}>
            {saving && <Spinner size="sm" className="mr-2" />}
            {saving ? 'Saving...' : buttonText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
