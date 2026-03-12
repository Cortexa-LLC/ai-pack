/**
 * Canonical role identifiers — matches the role names in roles/*.md on the server.
 * Use these constants instead of inline string literals throughout the GUI.
 */
export const ROLES = {
  ORCHESTRATOR: 'orchestrator',
  ENGINEER: 'engineer',
  REVIEWER: 'reviewer',
  ARCHITECT: 'architect',
  SPELUNKER: 'spelunker',
  TESTER: 'tester',
  INSPECTOR: 'inspector',
  ARCHAEOLOGIST: 'archaeologist',
} as const;

export type Role = (typeof ROLES)[keyof typeof ROLES];

/** Default role used for the chat panel. */
export const DEFAULT_CHAT_ROLE: Role = ROLES.ORCHESTRATOR;
