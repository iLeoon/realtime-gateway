import type { Conversation, Participant } from '@/lib/validation';

export const getInitials = (value: string): string =>
  value
    .trim()
    .split(/\s+/)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .slice(0, 2)
    .join('');

const buildParticipantName = (participants: Participant[], currentUserId?: string): string => {
  const others = participants.filter((participant) => participant.id !== currentUserId);
  const people = others.length > 0 ? others : participants;

  if (people.length === 0) {
    return 'Empty conversation';
  }

  const first = people[0];
  if (!first) {
    return 'Empty conversation';
  }

  if (people.length === 1) {
    return first.displayName;
  }

  const second = people[1];
  if (!second) {
    return first.displayName;
  }

  if (people.length === 2) {
    return `${first.displayName}, ${second.displayName}`;
  }

  return `${first.displayName}, ${second.displayName} +${people.length - 2}`;
};

export const getConversationTitle = (conversation: Conversation, currentUserId?: string): string => {
  if (conversation.conversationType === 'group-chat') {
    return buildParticipantName(conversation.participants, currentUserId);
  }

  const otherParticipant = conversation.participants.find((participant) => participant.id !== currentUserId);
  return otherParticipant?.displayName ?? buildParticipantName(conversation.participants, currentUserId);
};

export const formatMessageTime = (value: string): string =>
  new Date(value).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit'
  });
