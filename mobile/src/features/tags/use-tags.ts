import { useQuery } from '@tanstack/react-query';

import { listTags } from './api';

export function useTags(query = '') {
  return useQuery({
    queryKey: ['tags', query],
    queryFn: () => listTags(query),
  });
}
