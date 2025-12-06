import { ApiProperty } from '@nestjs/swagger';
import {
  IsEmail,
  IsString,
  Matches,
  MaxLength,
  MinLength,
} from 'class-validator';

export class CreateUserDto {
  @ApiProperty({
    example: 'user@example.com',
    description: 'Email пользователя',
  })
  @IsEmail(
    {},
    {
      message: 'Некорректный email',
    },
  )
  email: string;

  @ApiProperty({
    example: 'SecurePass123!',
    description: 'Пароль (минимум 8 символов)',
  })
  @IsString()
  @MinLength(8, { message: 'Пароль должен быть минимум 8 символов' })
  @MaxLength(50, { message: 'Пароль не должен превышать 50 символов' })
  @Matches(/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, {
    message: 'Пароль должен содержать заглавные, строчные буквы и цифры',
  })
  password: string;

  @ApiProperty({
    example: 'Иван Иванов',
    description: 'Имя пользователя',
  })
  @IsString()
  @MinLength(2, { message: 'Имя должно быть минимум 2 символа' })
  @MaxLength(50, { message: 'Имя не должно превышать 50 символов' })
  name: string;
}
