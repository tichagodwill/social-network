import type {Followers, User} from '$lib/types'

export function transformUser(data: any): User {
  return {
    id: data.id,
    email: data.email,
    firstName: data.first_name,
    lastName: data.last_name,
    dateOfBirth: new Date(data.date_of_birth).toLocaleDateString('en-US', { year: 'numeric', month: '2-digit', day: '2-digit' }),
    username: data.username,
    aboutMe: data.about_me,
    createdAt: data.created_at,
    isPrivate: data.is_private,
    avatar: data.avatar
  };
}

export function transformFollowers(data: any[]): Followers[] {
  if (!data) return [];
  return data.map(item => ({
    id: item.id,
    userId: item.userId,
    username: item.username,
    avatar: item.avatar,
    firstName: item.first_name,
    lastName: item.last_name,
  }));
}