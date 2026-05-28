import { Form } from 'antd'
import type { FormProps } from 'antd'
import type { ReactNode } from 'react'

export type UiFormProps<Values = unknown> = FormProps<Values>

export function UiForm<Values = unknown>({ children, ...rest }: UiFormProps<Values>) {
  return (
    <Form<Values> layout="vertical" {...rest}>
      {children as ReactNode}
    </Form>
  )
}
