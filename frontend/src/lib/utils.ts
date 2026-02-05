import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"
import * as yaml from 'js-yaml';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Format ConfigMap YAML for display:
 * 1. Remove kubectl.kubernetes.io/last-applied-configuration annotation
 * 2. Format data values with proper multiline strings
 */
export function formatConfigMapYaml(yamlStr: string): string {
  try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const doc = yaml.load(yamlStr) as any;
    if (!doc) return yamlStr;

    // Remove kubectl.kubernetes.io/last-applied-configuration annotation
    if (doc.metadata?.annotations) {
      delete doc.metadata.annotations['kubectl.kubernetes.io/last-applied-configuration'];
      // Remove empty annotations object
      if (Object.keys(doc.metadata.annotations).length === 0) {
        delete doc.metadata.annotations;
      }
    }

    // Serialize with proper multiline string formatting
    return yaml.dump(doc, {
      lineWidth: -1, // Don't wrap lines
      noRefs: true,  // Don't use YAML references
      quotingType: '"',
      forceQuotes: false,
    });
  } catch {
    // If parsing fails, return original
    return yamlStr;
  }
}

/**
 * Convert memory string to Mi (MiB)
 * Supports: Ki, Mi, Gi, Ti, K, M, G, T, bytes
 * Examples: "16Gi" -> "16384Mi", "4096Mi" -> "4096Mi", "1024Ki" -> "1Mi"
 */
export function formatMemory(value: string | undefined | null): string {
  if (!value) return '-';

  const str = value.toString().trim();

  // Parse number and unit
  const match = str.match(/^(\d+(?:\.\d+)?)\s*([A-Za-z]*)$/);
  if (!match) return value;

  const num = parseFloat(match[1]);
  const unit = match[2].toLowerCase();

  let miValue: number;

  switch (unit) {
    case 'ki':
      miValue = num / 1024;
      break;
    case 'mi':
      miValue = num;
      break;
    case 'gi':
      miValue = num * 1024;
      break;
    case 'ti':
      miValue = num * 1024 * 1024;
      break;
    case 'k':
      miValue = num / 1024;
      break;
    case 'm':
      miValue = num;
      break;
    case 'g':
      miValue = num * 1024;
      break;
    case 't':
      miValue = num * 1024 * 1024;
      break;
    case '':
      // Bytes
      miValue = num / (1024 * 1024);
      break;
    default:
      return value;
  }

  // Format the result
  if (miValue >= 1024) {
    return `${(miValue / 1024).toFixed(1)}Gi`;
  }
  return `${Math.round(miValue)}Mi`;
}

/**
 * Convert CPU string to millicores (m)
 * Supports: m (millicores), cores (no unit)
 * Examples: "4" -> "4000m", "800m" -> "800m"
 */
export function formatCpu(value: string | undefined | null): string {
  if (!value) return '-';

  const str = value.toString().trim();

  // Parse number and unit
  const match = str.match(/^(\d+(?:\.\d+)?)\s*([A-Za-z]*)$/);
  if (!match) return value;

  const num = parseFloat(match[1]);
  const unit = match[2].toLowerCase();

  if (unit === 'm') {
    return `${Math.round(num)}m`;
  } else if (unit === '' || unit === 'cores') {
    // Cores to millicores
    return `${Math.round(num * 1000)}m`;
  }

  return value;
}
