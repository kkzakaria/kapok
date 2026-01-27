/**
 * TypeScript SDK Usage Example
 * 
 * This example demonstrates how to use the auto-generated
 * Kapok TypeScript SDK to interact with your backend.
 * 
 * Prerequisites:
 * 1. Run `kapok dev` to start the server
 * 2. Run `kapok generate sdk` to generate the SDK
 * 3. Build the SDK: `cd sdk/typescript && npm install && npm run build`
 */

import { KapokClient } from './sdk/typescript';

// Initialize the client
const client = new KapokClient('http://localhost:8080/api');

async function main() {
  try {
    console.log('🚀 Kapok TypeScript SDK Example\n');

    // 1. Create a new post
    console.log('1. Creating a new post...');
    const newPost = await client.posts.create({
      title: 'Hello from TypeScript SDK!',
      content: 'This post was created using the auto-generated TypeScript SDK.',
      published: true,
      author: 'SDK Example',
    });
    console.log('✓ Created post:', newPost.id);

    // 2. List all posts
    console.log('\n2. Fetching all posts...');
    const allPosts = await client.posts.list({ limit: 10 });
    console.log(`✓ Found ${allPosts.length} posts`);

    // 3. Get a specific post
    console.log('\n3. Fetching post by ID...');
    const post = await client.posts.getById(newPost.id);
    console.log(`✓ Post #${post.id}: ${post.title}`);

    // 4. Update the post
    console.log('\n4. Updating the post...');
    const updatedPost = await client.posts.update(newPost.id, {
      title: 'Hello from TypeScript SDK! (Updated)',
    });
    console.log(`✓ Updated post: ${updatedPost.title}`);

    // 5. Create a comment on the post
    console.log('\n5. Adding a comment...');
    const comment = await client.comments.create({
      postId: newPost.id,
      author: 'TypeScript Fan',
      content: 'Great post! TypeScript SDKs are awesome!',
    });
    console.log(`✓ Created comment: ${comment.id}`);

    // 6. List comments for the post
    console.log('\n6. Fetching comments...');
    const comments = await client.comments.list();
    const postComments = comments.filter(c => c.postId === newPost.id);
    console.log(`✓ Found ${postComments.length} comments for this post`);

    // 7. Delete the post (will cascade delete comments)
    console.log('\n7. Deleting the post...');
    await client.posts.delete(newPost.id);
    console.log('✓ Post deleted successfully');

    console.log('\n✅ All operations completed successfully!');

  } catch (error) {
    console.error('❌ Error:', error);
    process.exit(1);
  }
}

// Run the example
main();
