import { useState, useEffect, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { authApi } from '@/api';
import { getClusters } from '@/config/clusters';
import { useAuth } from '@/hooks/use-auth';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { Layers } from 'lucide-react';

function getAuthApiServer(): string {
  const clusters = getClusters();
  if (clusters.length === 0) {
    throw new Error('No clusters configured');
  }
  return clusters[0].api_server;
}

export function LoginPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { login, isAuthenticated } = useAuth();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [feishuLoading, setFeishuLoading] = useState(false);

  // Prevent duplicate login requests (React StrictMode calls useEffect twice)
  const loginAttempted = useRef(false);

  // Check if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  // Handle OAuth callback
  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const storedState = sessionStorage.getItem('oauth_state');

    if (code && !loginAttempted.current) {
      loginAttempted.current = true;

      // Clear URL params immediately to prevent re-processing
      setSearchParams({}, { replace: true });

      // Validate state to prevent CSRF
      if (state && storedState && state !== storedState) {
        setError('Invalid state parameter. Please try again.');
        return;
      }

      handleFeishuCallback(code, state || undefined);
    }
  }, [searchParams, setSearchParams]);

  const handleFeishuCallback = async (code: string, state?: string) => {
    setLoading(true);
    setError(null);

    try {
      const response = await authApi.feishuLogin(getAuthApiServer(), code, state);
      login(response.token, response.user);
      sessionStorage.removeItem('oauth_state');
      navigate('/', { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const handleFeishuLogin = async () => {
    setFeishuLoading(true);
    setError(null);

    try {
      const config = await authApi.getFeishuConfig(getAuthApiServer());

      // Generate random state for CSRF protection
      const state = Math.random().toString(36).substring(2, 15);
      sessionStorage.setItem('oauth_state', state);

      // Build Feishu authorization URL
      const authUrl = new URL(config.authorize_url);
      authUrl.searchParams.set('client_id', config.app_id);
      authUrl.searchParams.set('redirect_uri', config.redirect_uri);
      authUrl.searchParams.set('response_type', 'code');
      authUrl.searchParams.set('state', state);

      // Redirect to Feishu
      window.location.href = authUrl.toString();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get login config');
      setFeishuLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center">
              <Layers className="w-10 h-10 text-primary" />
            </div>
          </div>
          <CardTitle className="text-2xl">K8S Manager</CardTitle>
          <CardDescription>
            Kubernetes cluster management platform
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading ? (
            <div className="flex flex-col items-center py-8">
              <Spinner size="lg" />
              <p className="mt-4 text-muted-foreground">Logging in...</p>
            </div>
          ) : (
            <>
              {error && (
                <div className="p-3 rounded-md bg-destructive/10 text-destructive text-sm">
                  {error}
                </div>
              )}
              <Button
                className="w-full h-12"
                onClick={handleFeishuLogin}
                disabled={feishuLoading}
              >
                {feishuLoading ? (
                  <>
                    <Spinner size="sm" className="mr-2" />
                    Redirecting...
                  </>
                ) : (
                  <>
                    <svg
                      className="w-5 h-5 mr-2"
                      viewBox="0 0 24 24"
                      fill="currentColor"
                    >
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1.41 14.59L7.17 13.17l1.41-1.41 2 2 5-5 1.41 1.41-6.41 6.42z" />
                    </svg>
                    Sign in with Feishu
                  </>
                )}
              </Button>
              <p className="text-center text-xs text-muted-foreground">
                Sign in with your company Feishu account
              </p>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
