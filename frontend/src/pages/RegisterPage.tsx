import * as React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowLeft, UserPlus, KeyRound } from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { register } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

export function RegisterPage() {
  const { login: authLogin, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [confirm, setConfirm] = React.useState('');
  const [inviteToken, setInviteToken] = React.useState('');
  const [error, setError] = React.useState('');
  const [isLoading, setIsLoading] = React.useState(false);

  React.useEffect(() => {
    if (isAuthenticated) {
      navigate('/servers', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!email.trim() || !password.trim()) {
      setError('Please fill in all fields.');
      return;
    }
    if (password !== confirm) {
      setError('Passwords do not match.');
      return;
    }
    if (password.length < 8) {
      setError('Password must be at least 8 characters.');
      return;
    }
    setIsLoading(true);
    try {
      await register(email.trim(), password, inviteToken.trim() || undefined);
      await authLogin();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <main className="relative flex min-h-screen flex-1 items-center justify-center overflow-hidden px-4 py-12 sm:px-6">
      <button type="button" onClick={() => navigate('/')} className="absolute left-5 top-5 inline-flex items-center gap-2 text-xs text-muted transition-colors hover:text-active sm:left-8 sm:top-8">
        <ArrowLeft className="h-3.5 w-3.5" />
        Back to Trace
      </button>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
        className="w-full min-w-0 max-w-sm"
      >
        <Card hover={false} className="border-white/12 bg-surface/90 p-6 shadow-[0_20px_70px_rgba(0,0,0,0.28)]">
          <div className="mb-6 text-center">
            <div className="mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-full border border-border bg-surface">
              <KeyRound className="h-5 w-5 text-accent" />
            </div>
            <p className="text-xs font-mono uppercase tracking-[0.16em] text-muted">Trace account</p>
            <h1 className="mt-2 text-xl font-bold tracking-tight text-active">Create account</h1>
            <p className="mt-1 text-sm text-muted">Create one workspace for every connected node.</p>
            <div className="mt-3 rounded-md border border-accent/20 bg-accent/5 px-3 py-2 font-mono text-[10px] uppercase tracking-wider text-accent">
              One account for all connected nodes
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs font-mono uppercase text-muted">Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
                autoFocus
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-mono uppercase text-muted">Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-mono uppercase text-muted">Confirm password</label>
              <input
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                placeholder="••••••••"
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-mono uppercase text-muted">
                Invite token <span className="text-muted/60">(optional)</span>
              </label>
              <input
                type="password"
                value={inviteToken}
                onChange={(e) => setInviteToken(e.target.value)}
                placeholder="Required when registration is disabled"
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
              />
              <p className="text-[10px] text-muted/70">
                Only needed for private demos where public registration is closed.
              </p>
            </div>

            {error && (
              <div className="rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-400">
                {error}
              </div>
            )}

            <Button variant="neon" size="md" type="submit" className="w-full gap-2" disabled={isLoading}>
              <UserPlus className="h-4 w-4" />
              {isLoading ? 'Creating account…' : 'Create account'}
            </Button>
          </form>

          <div className="mt-5 text-center text-xs text-muted">
            Already have an account?{' '}
            <Link to="/login" className="text-accent transition-opacity hover:opacity-80">
              Sign in
            </Link>
          </div>
        </Card>
      </motion.div>
    </main>
  );
}
