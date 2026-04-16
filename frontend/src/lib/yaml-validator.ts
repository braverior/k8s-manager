import yaml from 'js-yaml';

export interface YamlValidationResult {
  valid: boolean;
  error?: string;
}

/**
 * 验证 YAML 语法是否合法
 */
export function validateYamlSyntax(content: string): YamlValidationResult {
  if (!content.trim()) {
    return { valid: false, error: 'YAML 内容不能为空' };
  }
  try {
    yaml.load(content);
    return { valid: true };
  } catch (e) {
    if (e instanceof yaml.YAMLException) {
      return { valid: false, error: `YAML 语法错误: ${e.message.split('\n')[0]}` };
    }
    return { valid: false, error: 'YAML 解析失败' };
  }
}

/**
 * 验证 Kubernetes 资源 YAML（需要 apiVersion、kind、metadata.name）
 */
export function validateK8sYaml(content: string): YamlValidationResult {
  const syntaxResult = validateYamlSyntax(content);
  if (!syntaxResult.valid) return syntaxResult;

  try {
    const doc = yaml.load(content) as Record<string, unknown>;
    if (!doc || typeof doc !== 'object') {
      return { valid: false, error: 'YAML 内容不是有效的对象' };
    }
    if (!doc['apiVersion']) {
      return { valid: false, error: '缺少必填字段: apiVersion' };
    }
    if (!doc['kind']) {
      return { valid: false, error: '缺少必填字段: kind' };
    }
    const metadata = doc['metadata'] as Record<string, unknown> | undefined;
    if (!metadata || !metadata['name']) {
      return { valid: false, error: '缺少必填字段: metadata.name' };
    }
    return { valid: true };
  } catch {
    return { valid: false, error: 'YAML 结构解析失败' };
  }
}

/**
 * 验证 kubeconfig YAML（需要 apiVersion、clusters、contexts、users、current-context）
 */
export function validateKubeconfigYaml(content: string): YamlValidationResult {
  const syntaxResult = validateYamlSyntax(content);
  if (!syntaxResult.valid) return syntaxResult;

  try {
    const doc = yaml.load(content) as Record<string, unknown>;
    if (!doc || typeof doc !== 'object') {
      return { valid: false, error: 'kubeconfig 不是有效的 YAML 对象' };
    }
    const requiredFields = ['apiVersion', 'clusters', 'contexts', 'users'];
    for (const field of requiredFields) {
      if (!doc[field]) {
        return { valid: false, error: `kubeconfig 缺少必填字段: ${field}` };
      }
    }
    if (!Array.isArray(doc['clusters']) || (doc['clusters'] as unknown[]).length === 0) {
      return { valid: false, error: 'kubeconfig clusters 列表不能为空' };
    }
    return { valid: true };
  } catch {
    return { valid: false, error: 'kubeconfig 结构解析失败' };
  }
}
