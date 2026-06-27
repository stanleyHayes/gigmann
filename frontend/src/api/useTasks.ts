import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type Task = components['schemas']['Task']
export type TaskStatus = Task['status']
export type TaskCreate = components['schemas']['TaskCreate']

/** useTasks fetches the executive's "My Day" tasks. */
export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: async (): Promise<Task[]> => {
      const { data, error } = await api.GET('/api/v1/tasks')
      if (error || !data) {
        throw new Error('failed to load tasks')
      }
      return data.tasks
    },
  })
}

/** useUpdateTaskStatus moves a task between todo/in_progress/done. */
export function useUpdateTaskStatus() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (vars: { id: string; status: TaskStatus }): Promise<Task> => {
      const { data, error } = await api.POST('/api/v1/tasks/{taskId}/status', {
        params: { path: { taskId: vars.id } },
        body: { status: vars.status },
      })
      if (error || !data) {
        throw new Error('failed to update task')
      }
      return data
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['tasks'] })
    },
  })
}

/** useCreateTask creates a "My Day" task (e.g. from a brief item). */
export function useCreateTask() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (body: TaskCreate): Promise<Task> => {
      const { data, error } = await api.POST('/api/v1/tasks', { body })
      if (error || !data) {
        throw new Error('failed to create task')
      }
      return data
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['tasks'] })
    },
  })
}
