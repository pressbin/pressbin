-- Remove baked-in dev sample posts; content should match the GitHub content repo via sync.

DELETE FROM posts WHERE slug IN (
    'welcome-to-pressbin',
    'why-sqlite',
    'draft-notes'
);
