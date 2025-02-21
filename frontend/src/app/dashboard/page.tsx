'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import TaskList from '@/components/tasks/TaskList';
import CreateTaskForm from '@/components/tasks/CreateTaskForm';
import { createWebSocket } from '@/lib/api';

export default function DashboardPage() {
  const router = useRouter();
  const [showCreateForm, setShowCreateForm] = useState(false);

  useEffect(() => {
    // Check if user is authenticated
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    // Setup WebSocket connection
    const ws = createWebSocket();
    
    ws.onmessage = (event) => {
      // Handle real-time updates
      const data = JSON.parse(event.data);
      // You could update the task list here or show notifications
      console.log('WebSocket message:', data);
    };

    return () => {
      ws.close();
    };
  }, []);

  return (
    <div className="min-h-screen bg-gray-100">
      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="flex justify-between items-center mb-6">
            <h1 className="text-3xl font-bold text-gray-900">Tasks</h1>
            <button
              onClick={() => setShowCreateForm(!showCreateForm)}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              {showCreateForm ? 'Hide Form' : 'Create New Task'}
            </button>
          </div>

          {showCreateForm && (
            <div className="mb-6">
              <CreateTaskForm
                onTaskCreated={() => {
                  setShowCreateForm(false);
                  // The TaskList component will automatically refresh
                }}
              />
            </div>
          )}

          <TaskList />
        </div>
      </div>
    </div>
  );
}
