import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchTodos, fetchFacets, createTodo, updateTodo, deleteTodo } from '../api/todoApi';

// Custom hook to fetch all todos
export const useTodos = (searchQuery: string = '', status?: number, page: number = 1, limit: number = 10) => {
  return useQuery({
    queryKey: ['todos', searchQuery, status, page, limit],
    queryFn: () => fetchTodos(searchQuery, status, page, limit),
  });
};

export const useFacets = (searchQuery: string = '') => {
  return useQuery({
    queryKey: ['facets', searchQuery],
    queryFn: () => fetchFacets(searchQuery),
  });
};

// Custom hook to create a todo
export const useCreateTodo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createTodo,
    onSuccess: () => {
      // Invalidate and refetch the 'todos' and 'facets' queries so the UI updates automatically
      queryClient.invalidateQueries({ queryKey: ['todos'] });
      queryClient.invalidateQueries({ queryKey: ['facets'] });
    },
  });
};

// Custom hook to update a todo
export const useUpdateTodo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateTodo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
      queryClient.invalidateQueries({ queryKey: ['facets'] });
    },
  });
};

// Custom hook to delete a todo
export const useDeleteTodo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteTodo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
      queryClient.invalidateQueries({ queryKey: ['facets'] });
    },
  });
};
