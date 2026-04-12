// Redirect index to documents
// Uses shared design tokens from /shared/theme
import { useEffect } from 'react';
import { useRouter } from 'expo-router';

export default function AppIndexScreen() {
  const router = useRouter();

  useEffect(() => {
    router.replace('/(app)/documents');
  }, []);

  return null;
}
