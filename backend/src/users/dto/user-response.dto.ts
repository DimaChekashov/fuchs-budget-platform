import { Exclude } from 'class-transformer';

export class UserResponseDto {
  id: number;
  email: string;
  name: string;
  defaultCurrencyCode: string;
  createdAt: Date;
  updatedAt: Date;

  @Exclude()
  passwordHash: string;
}
