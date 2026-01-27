/**
 * React Hooks Usage Example
 * 
 * This example demonstrates how to use the auto-generated
 * Kapok React hooks in your React application.
 * 
 * Prerequisites:
 * 1. Run `kapok dev` to start the server
 * 2. Run `kapok generate react` to generate the hooks
 * 3. Install dependencies: npm install @tanstack/react-query
 */

import React, { useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { 
  KapokProvider, 
  useListPosts, 
  useCreatePosts,
  useDeletePosts,
  useListComments,
  useCreateComments 
} from 'kapok-react';

// Create a client
const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <KapokProvider baseUrl="http://localhost:8080/api">
        <div style={{ maxWidth: '800px', margin: '0 auto', padding: '20px' }}>
          <h1>🚀 Kapok React Hooks Example</h1>
          <BlogPosts />
        </div>
      </KapokProvider>
    </QueryClientProvider>
  );
}

function BlogPosts() {
  const [newPostTitle, setNewPostTitle] = useState('');
  const [newPostContent, setNewPostContent] = useState('');
  
  // Query hooks
  const { data: posts, isLoading, error } = useListPosts();
  const { data: comments } = useListComments();
  
  // Mutation hooks
  const createPost = useCreatePosts();
  const deletePost = useDeletePosts();
  const createComment = useCreateComments();

  const handleCreatePost = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!newPostTitle.trim()) return;

    await createPost.mutateAsync({
      title: newPostTitle,
      content: newPostContent,
      published: true,
      author: 'React User',
    });

    // Reset form
    setNewPostTitle('');
    setNewPostContent('');
  };

  const handleDeletePost = async (postId: number) => {
    if (window.confirm('Are you sure you want to delete this post?')) {
      await deletePost.mutateAsync(postId);
    }
  };

  const handleAddComment = async (postId: number) => {
    const content = window.prompt('Enter your comment:');
    if (content) {
      await createComment.mutateAsync({
        post_id: postId,
        author: 'Anonymous',
        content,
      });
    }
  };

  if (isLoading) {
    return <div style={{ textAlign: 'center', marginTop: '50px' }}>Loading posts...</div>;
  }

  if (error) {
    return (
      <div style={{ color: 'red', padding: '20px', border: '1px solid red', borderRadius: '4px' }}>
        <strong>Error loading posts:</strong> {error.message}
      </div>
    );
  }

  const getCommentsForPost = (postId: number) => {
    return comments?.filter(c => c.post_id === postId) || [];
  };

  return (
    <div>
      {/* Create Post Form */}
      <div style={{ 
        padding: '20px', 
        backgroundColor: '#f5f5f5', 
        borderRadius: '8px', 
        marginBottom: '30px' 
      }}>
        <h2>Create New Post</h2>
        <form onSubmit={handleCreatePost}>
          <input
            type="text"
            placeholder="Post title"
            value={newPostTitle}
            onChange={(e) => setNewPostTitle(e.target.value)}
            style={{ 
              width: '100%', 
              padding: '10px', 
              marginBottom: '10px',
              fontSize: '16px',
              border: '1px solid #ddd',
              borderRadius: '4px'
            }}
          />
          <textarea
            placeholder="Post content"
            value={newPostContent}
            onChange={(e) => setNewPostContent(e.target.value)}
            rows={3}
            style={{ 
              width: '100%', 
              padding: '10px', 
              marginBottom: '10px',
              fontSize: '16px',
              border: '1px solid #ddd',
              borderRadius: '4px'
            }}
          />
          <button
            type="submit"
            disabled={createPost.isPending}
            style={{
              padding: '10px 20px',
              fontSize: '16px',
              backgroundColor: '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: createPost.isPending ? 'wait' : 'pointer',
            }}
          >
            {createPost.isPending ? 'Creating...' : 'Create Post'}
          </button>
        </form>
      </div>

      {/* Posts List */}
      <h2>Blog Posts ({posts?.length || 0})</h2>
      {posts && posts.length === 0 ? (
        <p style={{ color: '#666', fontStyle: 'italic' }}>
          No posts yet. Create your first post above!
        </p>
      ) : (
        <div>
          {posts?.map((post) => (
            <article
              key={post.id}
              style={{
                padding: '20px',
                marginBottom: '20px',
                border: '1px solid #ddd',
                borderRadius: '8px',
                backgroundColor: 'white',
              }}
            >
              <h3 style={{ marginTop: 0 }}>{post.title}</h3>
              <p style={{ color: '#666' }}>{post.content}</p>
              <div style={{ fontSize: '12px', color: '#999', marginBottom: '10px' }}>
                By {post.author || 'Anonymous'} • {new Date(post.created_at).toLocaleDateString()}
              </div>
              
              {/* Comments */}
              <div style={{ 
                marginTop: '15px', 
                paddingTop: '15px', 
                borderTop: '1px solid #eee' 
              }}>
                <strong style={{ fontSize: '14px' }}>
                  Comments ({getCommentsForPost(post.id).length})
                </strong>
                <div style={{ marginTop: '10px' }}>
                  {getCommentsForPost(post.id).map((comment) => (
                    <div
                      key={comment.id}
                      style={{
                        padding: '8px',
                        marginTop: '8px',
                        backgroundColor: '#f9f9f9',
                        borderRadius: '4px',
                        fontSize: '14px',
                      }}
                    >
                      <strong>{comment.author}:</strong> {comment.content}
                    </div>
                  ))}
                </div>
              </div>

              {/* Actions */}
              <div style={{ marginTop: '15px', display: 'flex', gap: '10px' }}>
                <button
                  onClick={() => handleAddComment(post.id)}
                  style={{
                    padding: '6px 12px',
                    fontSize: '14px',
                    backgroundColor: '#28a745',
                    color: 'white',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  Add Comment
                </button>
                <button
                  onClick={() => handleDeletePost(post.id)}
                  disabled={deletePost.isPending}
                  style={{
                    padding: '6px 12px',
                    fontSize: '14px',
                    backgroundColor: '#dc3545',
                    color: 'white',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: deletePost.isPending ? 'wait' : 'pointer',
                  }}
                >
                  {deletePost.isPending ? 'Deleting...' : 'Delete Post'}
                </button>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}

export default App;
