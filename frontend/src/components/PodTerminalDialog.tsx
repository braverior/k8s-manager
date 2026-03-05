import { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import '@xterm/xterm/css/xterm.css';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Terminal as TerminalIcon, RefreshCw, X, Maximize2, Minimize2 } from 'lucide-react';
import type { Pod, PodContainer, TerminalMessage } from '@/types';

interface PodTerminalDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  pod: Pod | null;
  cluster: string;
  namespace: string;
}

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

export function PodTerminalDialog({
  open,
  onOpenChange,
  pod,
  cluster,
  namespace,
}: PodTerminalDialogProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstance = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const pingIntervalRef = useRef<number | null>(null);
  const isInitializedRef = useRef(false);

  const [selectedContainer, setSelectedContainer] = useState<string>('');
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [isFullscreen, setIsFullscreen] = useState(false);
  // Set default container when pod changes
  useEffect(() => {
    if (pod?.containers?.length) {
      const runningContainer = pod.containers.find((c) =>
        c.state?.toLowerCase() === 'running'
      );
      setSelectedContainer(runningContainer?.name || pod.containers[0].name);
    } else {
      setSelectedContainer('');
    }
  }, [pod]);

  // Initialize terminal when dialog opens
  useEffect(() => {
    if (!open) {
      // Cleanup when dialog closes
      if (wsRef.current) {
        wsRef.current.close(1000);
        wsRef.current = null;
      }
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current);
        pingIntervalRef.current = null;
      }
      if (terminalInstance.current) {
        terminalInstance.current.dispose();
        terminalInstance.current = null;
      }
      fitAddonRef.current = null;
      isInitializedRef.current = false;
      setConnectionStatus('disconnected');
      return;
    }

    // Initialize terminal after a short delay to ensure DOM is ready
    const initTimer = setTimeout(() => {
      if (!terminalRef.current || isInitializedRef.current) {
        console.log('[Terminal] Skip init: ref=', !!terminalRef.current, 'initialized=', isInitializedRef.current);
        return;
      }

      console.log('[Terminal] Initializing terminal...');
      isInitializedRef.current = true;

      const terminal = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Menlo, Monaco, "Courier New", monospace',
        theme: {
          background: '#1e1e1e',
          foreground: '#d4d4d4',
          cursor: '#d4d4d4',
          cursorAccent: '#1e1e1e',
          selectionBackground: '#264f78',
        },
      });

      const fitAddon = new FitAddon();
      terminal.loadAddon(fitAddon);
      terminal.loadAddon(new WebLinksAddon());

      terminal.open(terminalRef.current);

      // Fit after a small delay
      setTimeout(() => {
        fitAddon.fit();
        console.log('[Terminal] Terminal ready, cols:', terminal.cols, 'rows:', terminal.rows);
      }, 50);

      terminalInstance.current = terminal;
      fitAddonRef.current = fitAddon;

      // Handle terminal input
      terminal.onData((data) => {
        console.log('[Terminal] Input:', JSON.stringify(data));
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          const message: TerminalMessage = { type: 'input', data };
          wsRef.current.send(JSON.stringify(message));
        }
      });

      // Write initial message
      terminal.writeln('\x1b[90mTerminal ready. Click "Connect" to start.\x1b[0m');
      terminal.focus();
    }, 200);

    // Handle window resize
    const handleResize = () => {
      if (fitAddonRef.current && terminalInstance.current) {
        fitAddonRef.current.fit();
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          const message: TerminalMessage = {
            type: 'resize',
            cols: terminalInstance.current.cols,
            rows: terminalInstance.current.rows,
          };
          wsRef.current.send(JSON.stringify(message));
        }
      }
    };

    window.addEventListener('resize', handleResize);

    return () => {
      clearTimeout(initTimer);
      window.removeEventListener('resize', handleResize);
    };
  }, [open]);

  // Connect to WebSocket
  const connect = () => {
    if (!pod || !selectedContainer) return;

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    // Clear terminal and show connecting message
    if (terminalInstance.current) {
      terminalInstance.current.clear();
      terminalInstance.current.writeln('\x1b[33mConnecting to container...\x1b[0m');
    }

    setConnectionStatus('connecting');
    setErrorMessage('');

    // Get token from localStorage
    const token = localStorage.getItem('token');
    if (!token) {
      setConnectionStatus('error');
      setErrorMessage('Not authenticated');
      return;
    }

    // Build WebSocket URL using same-origin
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/v1/clusters/${cluster}/namespaces/${namespace}/pods/${pod.name}/exec?token=${encodeURIComponent(token)}&container=${encodeURIComponent(selectedContainer)}`;

    console.log('[Terminal] Connecting to:', wsUrl.replace(/token=[^&]+/, 'token=***'));
    if (terminalInstance.current) {
      terminalInstance.current.writeln(`\x1b[90m${wsUrl.replace(/token=[^&]+/, 'token=***')}\x1b[0m`);
    }

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('[Terminal] WebSocket connected');
      setConnectionStatus('connected');
      if (terminalInstance.current) {
        terminalInstance.current.clear();
        terminalInstance.current.writeln('\x1b[32mConnected to container: ' + selectedContainer + '\x1b[0m\r\n');

        // Send initial resize
        const message: TerminalMessage = {
          type: 'resize',
          cols: terminalInstance.current.cols,
          rows: terminalInstance.current.rows,
        };
        ws.send(JSON.stringify(message));

        // Focus terminal
        terminalInstance.current.focus();
      }

      // Start ping interval
      pingIntervalRef.current = window.setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'ping' }));
        }
      }, 30000);
    };

    ws.onmessage = (event) => {
      console.log('[Terminal] Message received:', event.data.substring(0, 100));
      try {
        const message: TerminalMessage = JSON.parse(event.data);
        switch (message.type) {
          case 'output':
            if (terminalInstance.current) {
              terminalInstance.current.write(message.data);
            }
            break;
          case 'error':
            setErrorMessage(message.data);
            if (terminalInstance.current) {
              terminalInstance.current.writeln(`\x1b[31mError: ${message.data}\x1b[0m`);
            }
            break;
          case 'pong':
            // Heartbeat response
            break;
        }
      } catch {
        // Non-JSON message, write directly
        if (terminalInstance.current) {
          terminalInstance.current.write(event.data);
        }
      }
    };

    ws.onerror = (event) => {
      console.error('[Terminal] WebSocket error:', event);
      setConnectionStatus('error');
      setErrorMessage('Connection failed');
      if (terminalInstance.current) {
        terminalInstance.current.writeln('\x1b[31mConnection error\x1b[0m');
      }
    };

    ws.onclose = (event) => {
      console.log('[Terminal] WebSocket closed:', event.code, event.reason);
      setConnectionStatus('disconnected');
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current);
        pingIntervalRef.current = null;
      }
      if (terminalInstance.current && event.code !== 1000) {
        terminalInstance.current.writeln(`\x1b[33mConnection closed (code: ${event.code})\x1b[0m`);
      }
    };
  };

  // Disconnect
  const disconnect = () => {
    if (wsRef.current) {
      wsRef.current.close(1000);
      wsRef.current = null;
    }
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
      pingIntervalRef.current = null;
    }
    setConnectionStatus('disconnected');
    if (terminalInstance.current) {
      terminalInstance.current.writeln('\x1b[33mDisconnected\x1b[0m');
    }
  };

  const getStatusBadge = () => {
    switch (connectionStatus) {
      case 'connected':
        return <Badge variant="success">Connected</Badge>;
      case 'connecting':
        return <Badge variant="warning">Connecting...</Badge>;
      case 'error':
        return <Badge variant="destructive">Error</Badge>;
      default:
        return <Badge variant="outline">Disconnected</Badge>;
    }
  };

  // Toggle fullscreen and refit terminal
  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
    // Refit terminal after state change
    setTimeout(() => {
      if (fitAddonRef.current && terminalInstance.current) {
        fitAddonRef.current.fit();
        // Send resize to backend if connected
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          const message: TerminalMessage = {
            type: 'resize',
            cols: terminalInstance.current.cols,
            rows: terminalInstance.current.rows,
          };
          wsRef.current.send(JSON.stringify(message));
        }
        terminalInstance.current.focus();
      }
    }, 100);
  };

  // Check if we have any containers at all
  const hasContainers = (pod?.containers?.length ?? 0) > 0;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        hideCloseButton
        className={`flex flex-col p-0 transition-all duration-200 ${
          isFullscreen
            ? 'max-w-full w-full h-full max-h-full rounded-none'
            : 'max-w-4xl h-[80vh]'
        }`}
      >
        <DialogHeader className="px-6 py-4 border-b shrink-0">
          <DialogTitle className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <TerminalIcon className="w-5 h-5 text-primary" />
              <span>Terminal - {pod?.name}</span>
              {getStatusBadge()}
            </div>
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="icon"
                onClick={toggleFullscreen}
                title={isFullscreen ? 'Exit fullscreen' : 'Fullscreen'}
              >
                {isFullscreen ? (
                  <Minimize2 className="w-4 h-4" />
                ) : (
                  <Maximize2 className="w-4 h-4" />
                )}
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => onOpenChange(false)}
                title="Close"
              >
                <X className="w-4 h-4" />
              </Button>
            </div>
          </DialogTitle>
        </DialogHeader>

        {/* Control Bar */}
        <div className="flex items-center gap-4 px-6 py-3 bg-muted/50 border-b">
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Container:</span>
            <Select
              value={selectedContainer}
              onValueChange={setSelectedContainer}
              disabled={connectionStatus === 'connected'}
            >
              <SelectTrigger className="w-[350px]">
                <SelectValue placeholder="Select container" />
              </SelectTrigger>
              <SelectContent>
                {pod?.containers?.map((container: PodContainer) => (
                  <SelectItem key={container.name} value={container.name}>
                    <div className="flex items-center justify-between w-full gap-4">
                      <span>{container.name}</span>
                      <Badge
                        variant={container.state?.toLowerCase() === 'running' ? 'success' : 'secondary'}
                        className="text-xs"
                      >
                        {container.state}
                      </Badge>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            {connectionStatus === 'disconnected' || connectionStatus === 'error' ? (
              <Button
                size="sm"
                onClick={connect}
                disabled={!selectedContainer || !hasContainers}
              >
                <TerminalIcon className="w-4 h-4 mr-1" />
                Connect
              </Button>
            ) : connectionStatus === 'connecting' ? (
              <Button size="sm" disabled>
                <RefreshCw className="w-4 h-4 mr-1 animate-spin" />
                Connecting...
              </Button>
            ) : (
              <Button size="sm" variant="destructive" onClick={disconnect}>
                <X className="w-4 h-4 mr-1" />
                Disconnect
              </Button>
            )}

            {connectionStatus === 'connected' && (
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  disconnect();
                  setTimeout(connect, 100);
                }}
              >
                <RefreshCw className="w-4 h-4 mr-1" />
                Reconnect
              </Button>
            )}
          </div>

          {errorMessage && (
            <span className="text-sm text-destructive ml-auto">{errorMessage}</span>
          )}
        </div>

        {/* Terminal Container */}
        <div
          className="flex-1 p-2 bg-[#1e1e1e] overflow-hidden"
          onClick={() => terminalInstance.current?.focus()}
        >
          {!hasContainers ? (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              <div className="text-center">
                <TerminalIcon className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>No containers available</p>
                <p className="text-sm mt-1">The pod must have at least one container</p>
              </div>
            </div>
          ) : (
            <div ref={terminalRef} className="h-full w-full" />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
