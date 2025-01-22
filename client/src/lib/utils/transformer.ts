import type { User } from '$lib/types'

export function transformUser(data: any): User {
  return {
    id: data.id,
    email: data.email,
    firstName: data.first_name,
    lastName: data.last_name,
    dateOfBirth: data.date_of_birth,
    username: data.username,
    aboutMe: data.about_me,
    createdAt: data.created_at,
    isPrivate: data.is_private,
    avatar: data.avatar
  };
}