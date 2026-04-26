import { useQuery } from '@tanstack/react-query'
import { iamApi } from '../api/iam'

export function useMe() {
  return useQuery({
    queryKey: ['me'],
    queryFn: iamApi.me,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })
}
